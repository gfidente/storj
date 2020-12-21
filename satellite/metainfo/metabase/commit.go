// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package metabase

import (
	"context"
	"database/sql"
	"errors"
	"time"

	pgxerrcode "github.com/jackc/pgerrcode"
	"github.com/zeebo/errs"

	"storj.io/common/storj"
	"storj.io/storj/private/dbutil/pgutil/pgerrcode"
	"storj.io/storj/private/dbutil/txutil"
	"storj.io/storj/private/tagsql"
)

var (
	// ErrInvalidRequest is used to indicate invalid requests.
	ErrInvalidRequest = errs.Class("metabase: invalid request")
	// ErrConflict is used to indicate conflict with the request.
	ErrConflict = errs.Class("metabase: conflict")
)

// BeginObjectNextVersion contains arguments necessary for starting an object upload.
type BeginObjectNextVersion struct {
	ObjectStream

	ExpiresAt              *time.Time
	ZombieDeletionDeadline *time.Time

	Encryption storj.EncryptionParameters
}

// BeginObjectNextVersion adds a pending object to the database, with automatically assigned version.
func (db *DB) BeginObjectNextVersion(ctx context.Context, opts BeginObjectNextVersion) (committed Version, err error) {
	defer mon.Task()(&ctx)(&err)

	if err := opts.ObjectStream.Verify(); err != nil {
		return -1, err
	}

	switch {
	case opts.Encryption.IsZero() || opts.Encryption.CipherSuite == storj.EncUnspecified:
		return -1, ErrInvalidRequest.New("Encryption is missing")
	case opts.Encryption.BlockSize <= 0:
		return -1, ErrInvalidRequest.New("Encryption.BlockSize is negative or zero")
	case opts.Version != NextVersion:
		return -1, ErrInvalidRequest.New("Version should be metabase.NextVersion")
	}

	row := db.db.QueryRow(ctx, `
		INSERT INTO objects (
			project_id, bucket_name, object_key, version, stream_id,
			expires_at, encryption,
			zombie_deletion_deadline
		) VALUES (
			$1, $2, $3,
				coalesce((
					SELECT version + 1
					FROM objects
					WHERE project_id = $1 AND bucket_name = $2 AND object_key = $3
					ORDER BY version DESC
					LIMIT 1
				), 1),
			$4, $5, $6,
			$7)
		RETURNING version
	`, opts.ProjectID, opts.BucketName, []byte(opts.ObjectKey), opts.StreamID,
		opts.ExpiresAt, encryptionParameters{&opts.Encryption},
		opts.ZombieDeletionDeadline)

	var v int64
	if err := row.Scan(&v); err != nil {
		return -1, Error.New("unable to insert object: %w", err)
	}

	return Version(v), nil
}

// BeginObjectExactVersion contains arguments necessary for starting an object upload.
type BeginObjectExactVersion struct {
	ObjectStream

	ExpiresAt              *time.Time
	ZombieDeletionDeadline *time.Time

	Encryption storj.EncryptionParameters
}

// BeginObjectExactVersion adds a pending object to the database, with specific version.
func (db *DB) BeginObjectExactVersion(ctx context.Context, opts BeginObjectExactVersion) (committed Version, err error) {
	defer mon.Task()(&ctx)(&err)

	if err := opts.ObjectStream.Verify(); err != nil {
		return -1, err
	}

	switch {
	case opts.Encryption.IsZero() || opts.Encryption.CipherSuite == storj.EncUnspecified:
		return -1, ErrInvalidRequest.New("Encryption is missing")
	case opts.Encryption.BlockSize <= 0:
		return -1, ErrInvalidRequest.New("Encryption.BlockSize is negative or zero")
	case opts.Version == NextVersion:
		return -1, ErrInvalidRequest.New("Version should not be metabase.NextVersion")
	}

	_, err = db.db.ExecContext(ctx, `
		INSERT INTO objects (
			project_id, bucket_name, object_key, version, stream_id,
			expires_at, encryption,
			zombie_deletion_deadline
		) values (
			$1, $2, $3, $4, $5,
			$6, $7,
			$8
		)
	`,
		opts.ProjectID, opts.BucketName, []byte(opts.ObjectKey), opts.Version, opts.StreamID,
		opts.ExpiresAt, encryptionParameters{&opts.Encryption},
		opts.ZombieDeletionDeadline)
	if err != nil {
		if code := pgerrcode.FromError(err); code == pgxerrcode.UniqueViolation {
			return -1, ErrConflict.New("object already exists")
		}
		return -1, Error.New("unable to insert object: %w", err)
	}

	return opts.Version, nil
}

// BeginSegment contains options to verify, whether a new segment upload can be started.
type BeginSegment struct {
	ObjectStream

	Position    SegmentPosition
	RootPieceID storj.PieceID
	Pieces      Pieces
}

// BeginSegment verifies, whether a new segment upload can be started.
func (db *DB) BeginSegment(ctx context.Context, opts BeginSegment) (err error) {
	defer mon.Task()(&ctx)(&err)

	if err := opts.ObjectStream.Verify(); err != nil {
		return err
	}

	if err := opts.Pieces.Verify(); err != nil {
		return err
	}

	if opts.RootPieceID.IsZero() {
		return ErrInvalidRequest.New("RootPieceID missing")
	}

	// NOTE: this isn't strictly necessary, since we can also fail this in CommitSegment.
	//       however, we should prevent creating segements for non-partial objects.

	// NOTE: these queries could be combined into one.

	return txutil.WithTx(ctx, db.db, nil, func(ctx context.Context, tx tagsql.Tx) (err error) {
		// Verify that object exists and is partial.
		var value int
		err = tx.QueryRow(ctx, `
			SELECT 1
			FROM objects WHERE
				project_id   = $1 AND
				bucket_name  = $2 AND
				object_key   = $3 AND
				version      = $4 AND
				stream_id    = $5 AND
				status       = `+pendingStatus,
			opts.ProjectID, opts.BucketName, []byte(opts.ObjectKey), opts.Version, opts.StreamID).Scan(&value)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return Error.New("pending object missing")
			}
			return Error.New("unable to query object status: %w", err)
		}

		// Verify that the segment does not exist.
		err = tx.QueryRow(ctx, `
			SELECT 1
			FROM segments WHERE
				stream_id = $1 AND
				position  = $2
		`, opts.StreamID, opts.Position).Scan(&value)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return Error.New("unable to query segments: %w", err)
		}
		err = nil //nolint: ineffassign, ignore any other err result (explicitly)

		return nil
	})
}

// CommitSegment contains all necessary information about the segment.
type CommitSegment struct {
	ObjectStream

	Position    SegmentPosition
	RootPieceID storj.PieceID

	EncryptedKeyNonce []byte
	EncryptedKey      []byte

	PlainOffset   int64 // offset in the original data stream
	PlainSize     int32 // size before encryption
	EncryptedSize int32 // segment size after encryption

	Redundancy storj.RedundancyScheme

	Pieces Pieces
}

// CommitSegment commits segment to the database.
func (db *DB) CommitSegment(ctx context.Context, opts CommitSegment) (err error) {
	defer mon.Task()(&ctx)(&err)

	if err := opts.ObjectStream.Verify(); err != nil {
		return err
	}

	if err := opts.Pieces.Verify(); err != nil {
		return err
	}

	switch {
	case opts.RootPieceID.IsZero():
		return ErrInvalidRequest.New("RootPieceID missing")
	case len(opts.EncryptedKey) == 0:
		return ErrInvalidRequest.New("EncryptedKey missing")
	case len(opts.EncryptedKeyNonce) == 0:
		return ErrInvalidRequest.New("EncryptedKeyNonce missing")
	case opts.EncryptedSize <= 0:
		return ErrInvalidRequest.New("EncryptedSize negative or zero")
	case opts.PlainSize <= 0:
		return ErrInvalidRequest.New("PlainSize negative or zero")
	case opts.PlainOffset < 0:
		return ErrInvalidRequest.New("PlainOffset negative")
	case opts.Redundancy.IsZero():
		return ErrInvalidRequest.New("Redundancy zero")
	}

	// TODO: verify opts.Pieces is compatible with opts.Redundancy

	return txutil.WithTx(ctx, db.db, nil, func(ctx context.Context, tx tagsql.Tx) error {
		// Verify that object exists and is partial.
		var value int
		err = tx.QueryRowContext(ctx, `
			SELECT 1
			FROM objects WHERE
				project_id   = $1 AND
				bucket_name  = $2 AND
				object_key   = $3 AND
				version      = $4 AND
				stream_id    = $5 AND
				status       = `+pendingStatus,
			opts.ProjectID, opts.BucketName, []byte(opts.ObjectKey), opts.Version, opts.StreamID).Scan(&value)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return Error.New("pending object missing")
			}
			return Error.New("unable to query object status: %w", err)
		}

		// Insert into segments.
		_, err = tx.ExecContext(ctx, `
			INSERT INTO segments (
				stream_id, position,
				root_piece_id, encrypted_key_nonce, encrypted_key,
				encrypted_size, plain_offset, plain_size,
				redundancy,
				remote_pieces
			) VALUES (
				$1, $2,
				$3, $4, $5,
				$6, $7, $8,
				$9,
				$10
			)`,
			opts.StreamID, opts.Position,
			opts.RootPieceID, opts.EncryptedKeyNonce, opts.EncryptedKey,
			opts.EncryptedSize, opts.PlainOffset, opts.PlainSize,
			redundancyScheme{&opts.Redundancy},
			opts.Pieces,
		)
		if err != nil {
			if code := pgerrcode.FromError(err); code == pgxerrcode.UniqueViolation {
				return ErrConflict.New("segment already exists")
			}
			return Error.New("unable to insert segment: %w", err)
		}

		return nil
	})
}

// CommitInlineSegment contains all necessary information about the segment.
type CommitInlineSegment struct {
	ObjectStream

	Position SegmentPosition

	EncryptedKeyNonce []byte
	EncryptedKey      []byte

	PlainOffset int64 // offset in the original data stream
	PlainSize   int32 // size before encryption

	InlineData []byte
}

// CommitInlineSegment commits inline segment to the database.
func (db *DB) CommitInlineSegment(ctx context.Context, opts CommitInlineSegment) (err error) {
	defer mon.Task()(&ctx)(&err)

	if err := opts.ObjectStream.Verify(); err != nil {
		return err
	}

	// TODO: do we have a lower limit for inline data?

	switch {
	case len(opts.InlineData) == 0:
		return ErrInvalidRequest.New("InlineData missing")
	case len(opts.EncryptedKey) == 0:
		return ErrInvalidRequest.New("EncryptedKey missing")
	case len(opts.EncryptedKeyNonce) == 0:
		return ErrInvalidRequest.New("EncryptedKeyNonce missing")
	case opts.PlainSize <= 0:
		return ErrInvalidRequest.New("PlainSize negative or zero")
	case opts.PlainOffset < 0:
		return ErrInvalidRequest.New("PlainOffset negative")
	}

	return txutil.WithTx(ctx, db.db, nil, func(ctx context.Context, tx tagsql.Tx) error {
		// Verify that object exists and is partial.
		var value int
		err = tx.QueryRowContext(ctx, `
			SELECT 1
			FROM objects WHERE
				project_id   = $1 AND
				bucket_name  = $2 AND
				object_key   = $3 AND
				version      = $4 AND
				stream_id    = $5 AND
				status       = `+pendingStatus,
			opts.ProjectID, opts.BucketName, []byte(opts.ObjectKey), opts.Version, opts.StreamID).Scan(&value)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return Error.New("pending object missing")
			}
			return Error.New("unable to query object status: %w", err)
		}

		// Insert into segments.
		_, err = tx.ExecContext(ctx, `
			INSERT INTO segments (
				stream_id, position,
				root_piece_id, encrypted_key_nonce, encrypted_key,
				encrypted_size, plain_offset, plain_size,
				inline_data
			) VALUES (
				$1, $2,
				$3, $4, $5,
				$6, $7, $8,
				$9
			)`,
			opts.StreamID, opts.Position,
			storj.PieceID{}, opts.EncryptedKeyNonce, opts.EncryptedKey,
			len(opts.InlineData), opts.PlainOffset, opts.PlainSize,
			opts.InlineData,
		)
		if err != nil {
			if code := pgerrcode.FromError(err); code == pgxerrcode.UniqueViolation {
				return ErrConflict.New("segment already exists")
			}
			return Error.New("unable to insert segment: %w", err)
		}

		return nil
	})
}

// CommitObject contains arguments necessary for committing an object.
type CommitObject struct {
	ObjectStream

	EncryptedMetadata             []byte
	EncryptedMetadataNonce        []byte
	EncryptedMetadataEncryptedKey []byte
}

// CommitObject adds a pending object to the database.
func (db *DB) CommitObject(ctx context.Context, opts CommitObject) (object Object, err error) {
	defer mon.Task()(&ctx)(&err)

	if err := opts.ObjectStream.Verify(); err != nil {
		return Object{}, err
	}

	err = txutil.WithTx(ctx, db.db, nil, func(ctx context.Context, tx tagsql.Tx) error {
		type Segment struct {
			Position      SegmentPosition
			EncryptedSize int32
			PlainOffset   int64
			PlainSize     int32
		}

		var segments []Segment
		err = withRows(tx.Query(ctx, `
			SELECT position, encrypted_size, plain_offset, plain_size
			FROM segments
			WHERE stream_id = $1
			ORDER BY position
		`, opts.StreamID))(func(rows tagsql.Rows) error {
			for rows.Next() {
				var segment Segment
				err := rows.Scan(&segment.Position, &segment.EncryptedSize, &segment.PlainOffset, &segment.PlainSize)
				if err != nil {
					return Error.New("failed to scan segments: %w", err)
				}
				segments = append(segments, segment)
			}
			return nil
		})
		if err != nil {
			return Error.New("failed to fetch segments: %w", err)
		}

		// TODO disabled for now
		// verify segments
		// if len(segments) > 0 {
		// 	// without proofs we expect the segments to be contiguous
		// 	hasOffset := false
		// 	offset := int64(0)
		// 	for i, seg := range segments {
		// 		if seg.Position.Part != 0 && seg.Position.Index != uint32(i) {
		// 			return Error.New("expected segment (%d,%d), found (%d,%d)", 0, i, seg.Position.Part, seg.Position.Index)
		// 		}
		// 		if seg.PlainOffset != 0 {
		// 			hasOffset = true
		// 		}
		// 		if hasOffset && seg.PlainOffset != offset {
		// 			return Error.New("segment %d should be at plain offset %d, offset is %d", seg.Position.Index, offset, seg.PlainOffset)
		// 		}
		// 		offset += int64(seg.PlainSize)
		// 	}
		// }

		// TODO: would we even need this when we make main index plain_offset?
		fixedSegmentSize := int32(0)
		if len(segments) > 0 {
			fixedSegmentSize = segments[0].EncryptedSize
			for _, seg := range segments[:len(segments)-1] {
				if seg.EncryptedSize != fixedSegmentSize {
					fixedSegmentSize = -1
					break
				}
			}
		}

		var totalPlainSize, totalEncryptedSize int64
		for _, seg := range segments {
			totalPlainSize += int64(seg.PlainSize)
			totalEncryptedSize += int64(seg.EncryptedSize)
		}

		err = tx.QueryRow(ctx, `
			UPDATE objects SET
				status =`+committedStatus+`,
				segment_count = $6,

				encrypted_metadata_nonce         = $7,
				encrypted_metadata               = $8,
				encrypted_metadata_encrypted_key = $9,

				total_plain_size     = $10,
				total_encrypted_size = $11,
				fixed_segment_size   = $12,
				zombie_deletion_deadline = NULL
			WHERE
				project_id   = $1 AND
				bucket_name  = $2 AND
				object_key   = $3 AND
				version      = $4 AND
				stream_id    = $5 AND
				status       = `+pendingStatus+`
			RETURNING
				created_at, expires_at,
				encryption;
		`, opts.ProjectID, opts.BucketName, []byte(opts.ObjectKey), opts.Version, opts.StreamID,
			len(segments),
			opts.EncryptedMetadataNonce, opts.EncryptedMetadata, opts.EncryptedMetadataEncryptedKey,
			totalPlainSize,
			totalEncryptedSize,
			fixedSegmentSize,
		).
			Scan(
				&object.CreatedAt, &object.ExpiresAt,
				encryptionParameters{&object.Encryption},
			)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return storj.ErrObjectNotFound.Wrap(Error.New("object with specified version and pending status is missing"))
			}
			return Error.New("failed to update object: %w", err)
		}

		object.StreamID = opts.StreamID
		object.ProjectID = opts.ProjectID
		object.BucketName = opts.BucketName
		object.ObjectKey = opts.ObjectKey
		object.Version = opts.Version
		object.Status = Committed
		object.SegmentCount = int32(len(segments))
		object.EncryptedMetadataNonce = opts.EncryptedMetadataNonce
		object.EncryptedMetadata = opts.EncryptedMetadata
		object.EncryptedMetadataEncryptedKey = opts.EncryptedMetadataEncryptedKey
		object.TotalPlainSize = totalPlainSize
		object.TotalEncryptedSize = totalEncryptedSize
		object.FixedSegmentSize = fixedSegmentSize
		return nil
	})
	if err != nil {
		return Object{}, err
	}
	return object, nil
}

// UpdateObjectMetadata contains arguments necessary for updating an object metadata.
type UpdateObjectMetadata struct {
	ObjectStream

	EncryptedMetadata             []byte
	EncryptedMetadataNonce        []byte
	EncryptedMetadataEncryptedKey []byte
}

// UpdateObjectMetadata updates an object metadata.
func (db *DB) UpdateObjectMetadata(ctx context.Context, opts UpdateObjectMetadata) (err error) {
	defer mon.Task()(&ctx)(&err)

	if err := opts.ObjectStream.Verify(); err != nil {
		return err
	}

	if opts.ObjectStream.Version <= 0 {
		return ErrInvalidRequest.New("Version invalid: %v", opts.Version)
	}

	// TODO So the issue is that during a multipart upload of an object,
	// uplink can update object metadata. If we add the arguments EncryptedMetadata
	// to CommitObject, they will need to account for them being optional.
	// Leading to scenarios where uplink calls update metadata, but wants to clear them
	// during commit object.
	result, err := db.db.ExecContext(ctx, `
		UPDATE objects SET
			encrypted_metadata_nonce         = $6,
			encrypted_metadata               = $7,
			encrypted_metadata_encrypted_key = $8
		WHERE
			project_id   = $1 AND
			bucket_name  = $2 AND
			object_key   = $3 AND
			version      = $4 AND
			stream_id    = $5 AND
			status       = `+committedStatus,
		opts.ProjectID, opts.BucketName, []byte(opts.ObjectKey), opts.Version, opts.StreamID,
		opts.EncryptedMetadataNonce, opts.EncryptedMetadata, opts.EncryptedMetadataEncryptedKey)
	if err != nil {
		return Error.New("unable to update object metadata: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return Error.New("failed to get rows affected: %w", err)
	}

	if affected == 0 {
		return storj.ErrObjectNotFound.Wrap(
			Error.New("object with specified version and committed status is missing"),
		)
	}

	return nil
}
