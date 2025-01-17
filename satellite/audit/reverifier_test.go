// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package audit_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"storj.io/common/memory"
	"storj.io/common/pb"
	"storj.io/common/storj"
	"storj.io/common/testcontext"
	"storj.io/common/testrand"
	"storj.io/storj/private/testplanet"
	"storj.io/storj/satellite/audit"
	"storj.io/storj/satellite/metabase"
)

func TestReverifyPiece(t *testing.T) {
	testplanet.Run(t, testplanet.Config{
		SatelliteCount:   1,
		StorageNodeCount: 5,
		UplinkCount:      1,
		Reconfigure: testplanet.Reconfigure{
			Satellite: testplanet.ReconfigureRS(3, 4, 4, 5),
		},
	}, func(t *testing.T, ctx *testcontext.Context, planet *testplanet.Planet) {
		satellite := planet.Satellites[0]
		audits := satellite.Audit

		audits.Worker.Loop.Pause()
		audits.Chore.Loop.Pause()

		segment := uploadSomeData(t, ctx, planet)

		// ensure ReverifyPiece tells us to remove the segment from the queue after a successful audit
		for _, piece := range segment.Pieces {
			keepInQueue := satellite.Audit.Verifier.ReverifyPiece(ctx, audit.PieceLocator{
				StreamID: segment.StreamID,
				Position: segment.Position,
				NodeID:   piece.StorageNode,
				PieceNum: int(piece.Number),
			})
			require.False(t, keepInQueue)
		}
	})
}

func TestDoReverifyPieceSucceeds(t *testing.T) {
	testplanet.Run(t, testplanet.Config{
		SatelliteCount:   1,
		StorageNodeCount: 5,
		UplinkCount:      1,
		Reconfigure: testplanet.Reconfigure{
			Satellite: testplanet.ReconfigureRS(3, 4, 4, 5),
		},
	}, func(t *testing.T, ctx *testcontext.Context, planet *testplanet.Planet) {
		satellite := planet.Satellites[0]
		audits := satellite.Audit

		audits.Worker.Loop.Pause()
		audits.Chore.Loop.Pause()

		segment := uploadSomeData(t, ctx, planet)

		// ensure DoReverifyPiece succeeds on the new pieces we uploaded
		for _, piece := range segment.Pieces {
			outcome, err := satellite.Audit.Verifier.DoReverifyPiece(ctx, planet.Log().Named("reverifier"), audit.PieceLocator{
				StreamID: segment.StreamID,
				Position: segment.Position,
				NodeID:   piece.StorageNode,
				PieceNum: int(piece.Number),
			})
			require.NoError(t, err)
			require.Equal(t, audit.OutcomeSuccess, outcome)
		}
	})
}

func TestDoReverifyPieceWithNodeOffline(t *testing.T) {
	testplanet.Run(t, testplanet.Config{
		SatelliteCount:   1,
		StorageNodeCount: 5,
		UplinkCount:      1,
		Reconfigure: testplanet.Reconfigure{
			Satellite: testplanet.ReconfigureRS(3, 4, 4, 5),
		},
	}, func(t *testing.T, ctx *testcontext.Context, planet *testplanet.Planet) {
		satellite := planet.Satellites[0]
		audits := satellite.Audit

		audits.Worker.Loop.Pause()
		audits.Chore.Loop.Pause()

		segment := uploadSomeData(t, ctx, planet)

		offlinePiece := segment.Pieces[0]
		offlineNode := planet.FindNode(offlinePiece.StorageNode)
		require.NotNil(t, offlineNode)
		require.NoError(t, planet.StopPeer(offlineNode))

		// see what happens when DoReverifyPiece tries to hit that node
		outcome, err := satellite.Audit.Verifier.DoReverifyPiece(ctx, planet.Log().Named("reverifier"), audit.PieceLocator{
			StreamID: segment.StreamID,
			Position: segment.Position,
			NodeID:   offlinePiece.StorageNode,
			PieceNum: int(offlinePiece.Number),
		})
		require.NoError(t, err)
		require.Equal(t, audit.OutcomeNodeOffline, outcome)
	})
}

func TestDoReverifyPieceWithPieceMissing(t *testing.T) {
	testplanet.Run(t, testplanet.Config{
		SatelliteCount:   1,
		StorageNodeCount: 5,
		UplinkCount:      1,
		Reconfigure: testplanet.Reconfigure{
			Satellite: testplanet.ReconfigureRS(3, 4, 4, 5),
		},
	}, func(t *testing.T, ctx *testcontext.Context, planet *testplanet.Planet) {
		satellite := planet.Satellites[0]
		audits := satellite.Audit

		audits.Worker.Loop.Pause()
		audits.Chore.Loop.Pause()

		segment := uploadSomeData(t, ctx, planet)

		// delete piece on a storage node
		missingPiece := segment.Pieces[0]
		missingPieceNode := planet.FindNode(missingPiece.StorageNode)
		missingPieceID := segment.RootPieceID.Derive(missingPiece.StorageNode, int32(missingPiece.Number))
		err := missingPieceNode.Storage2.Store.Delete(ctx, satellite.ID(), missingPieceID)
		require.NoError(t, err)

		// see what happens when DoReverifyPiece tries to hit that node
		outcome, err := satellite.Audit.Verifier.DoReverifyPiece(ctx, planet.Log().Named("reverifier"), audit.PieceLocator{
			StreamID: segment.StreamID,
			Position: segment.Position,
			NodeID:   missingPiece.StorageNode,
			PieceNum: int(missingPiece.Number),
		})
		require.NoError(t, err)
		require.Equal(t, audit.OutcomeFailure, outcome)
	})
}

// This pattern is used for several tests.
func testReverifyRewrittenPiece(t *testing.T, mutator func(content []byte, header *pb.PieceHeader), expectedOutcome audit.Outcome) {
	testplanet.Run(t, testplanet.Config{
		SatelliteCount:   1,
		StorageNodeCount: 5,
		UplinkCount:      1,
		Reconfigure: testplanet.Reconfigure{
			Satellite: testplanet.ReconfigureRS(3, 4, 4, 5),
		},
	}, func(t *testing.T, ctx *testcontext.Context, planet *testplanet.Planet) {
		satellite := planet.Satellites[0]
		audits := satellite.Audit

		audits.Worker.Loop.Pause()
		audits.Chore.Loop.Pause()

		segment := uploadSomeData(t, ctx, planet)

		// rewrite piece on a storage node
		pieceToRewrite := segment.Pieces[0]
		node := planet.FindNode(pieceToRewrite.StorageNode)
		pieceID := segment.RootPieceID.Derive(pieceToRewrite.StorageNode, int32(pieceToRewrite.Number))

		rewritePiece(t, ctx, node, satellite.ID(), pieceID, mutator)

		outcome, err := satellite.Audit.Verifier.DoReverifyPiece(ctx, planet.Log().Named("reverifier"), audit.PieceLocator{
			StreamID: segment.StreamID,
			Position: segment.Position,
			NodeID:   pieceToRewrite.StorageNode,
			PieceNum: int(pieceToRewrite.Number),
		})
		require.NoError(t, err)
		require.Equal(t, expectedOutcome, outcome)
	})
}

// Some following tests rely on our ability to rewrite a piece on a storage node
// with new contents or a new piece header. Since the api for dealing with the
// piece store may change over time, this test makes sure that we can even expect
// the rewriting trick to work.
func TestDoReverifyPieceWithRewrittenPiece(t *testing.T) {
	testReverifyRewrittenPiece(t, func(content []byte, header *pb.PieceHeader) {
		// don't change anything; just write back original contents
	}, audit.OutcomeSuccess)
}

func TestDoReverifyPieceWithCorruptedContent(t *testing.T) {
	testReverifyRewrittenPiece(t, func(content []byte, header *pb.PieceHeader) {
		// increment last byte of content
		content[len(content)-1]++
	}, audit.OutcomeFailure)
}

func TestDoReverifyPieceWithCorruptedHash(t *testing.T) {
	testReverifyRewrittenPiece(t, func(content []byte, header *pb.PieceHeader) {
		// change last byte of hash
		header.Hash[len(header.Hash)-1]++
	}, audit.OutcomeFailure)
}

func TestDoReverifyPieceWithInvalidHashSignature(t *testing.T) {
	testReverifyRewrittenPiece(t, func(content []byte, header *pb.PieceHeader) {
		// change last byte of signature on hash
		header.Signature[len(header.Signature)-1]++
	}, audit.OutcomeFailure)
}

func TestDoReverifyPieceWithInvalidOrderLimitSignature(t *testing.T) {
	testReverifyRewrittenPiece(t, func(content []byte, header *pb.PieceHeader) {
		// change last byte of signature on order limit signature
		header.OrderLimit.SatelliteSignature[len(header.OrderLimit.SatelliteSignature)-1]++
	}, audit.OutcomeFailure)
}

func uploadSomeData(t *testing.T, ctx *testcontext.Context, planet *testplanet.Planet) metabase.Segment {
	satellite := planet.Satellites[0]

	// upload some random data
	ul := planet.Uplinks[0]
	testData := testrand.Bytes(8 * memory.KiB)
	err := ul.Upload(ctx, satellite, "audittest", "audit/test/path", testData)
	require.NoError(t, err)

	// of course, this way of getting the *metabase.Segment only works if
	// it was the first segment uploaded
	segments, err := satellite.Metabase.DB.TestingAllSegments(ctx)
	require.NoError(t, err)

	return segments[0]
}

func rewritePiece(t *testing.T, ctx *testcontext.Context, node *testplanet.StorageNode, satelliteID storj.NodeID, pieceID storj.PieceID, mutator func(contents []byte, header *pb.PieceHeader)) {
	reader, err := node.Storage2.Store.Reader(ctx, satelliteID, pieceID)
	require.NoError(t, err)
	pieceHeader, err := reader.GetPieceHeader()
	require.NoError(t, err)
	pieceContents, err := io.ReadAll(reader)
	require.NoError(t, err)
	err = reader.Close()
	require.NoError(t, err)

	mutator(pieceContents, pieceHeader)

	writer, err := node.Storage2.Store.Writer(ctx, satelliteID, pieceID, pieceHeader.HashAlgorithm)
	require.NoError(t, err)
	n, err := writer.Write(pieceContents)
	require.NoError(t, err)
	require.Equal(t, len(pieceContents), n)
	err = writer.Commit(ctx, pieceHeader)
	require.NoError(t, err)
}
