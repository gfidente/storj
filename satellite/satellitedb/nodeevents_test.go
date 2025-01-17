// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package satellitedb_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"storj.io/common/testcontext"
	"storj.io/common/uuid"
	"storj.io/storj/private/teststorj"
	"storj.io/storj/satellite"
	"storj.io/storj/satellite/nodeevents"
	"storj.io/storj/satellite/satellitedb/satellitedbtest"
)

func TestNodeEvents(t *testing.T) {
	satellitedbtest.Run(t, func(ctx *testcontext.Context, t *testing.T, db satellite.DB) {
		testID := teststorj.NodeIDFromString("test")
		testEmail := "test@storj.test"
		eventType := nodeevents.Disqualified

		neFromInsert, err := db.NodeEvents().Insert(ctx, testEmail, testID, eventType)
		require.NoError(t, err)
		require.NotNil(t, neFromInsert.ID)
		require.Equal(t, testID, neFromInsert.NodeID)
		require.Equal(t, testEmail, neFromInsert.Email)
		require.Equal(t, eventType, neFromInsert.Event)
		require.NotNil(t, neFromInsert.CreatedAt)
		require.Nil(t, neFromInsert.EmailSent)

		neFromGet, err := db.NodeEvents().GetLatestByEmailAndEvent(ctx, neFromInsert.Email, neFromInsert.Event)
		require.NoError(t, err)
		require.Equal(t, neFromInsert, neFromGet)
	})
}

func TestNodeEventsUpdateEmailSent(t *testing.T) {
	satellitedbtest.Run(t, func(ctx *testcontext.Context, t *testing.T, db satellite.DB) {
		testID1 := teststorj.NodeIDFromString("test1")
		testID2 := teststorj.NodeIDFromString("test2")
		testEmail1 := "test1@storj.test"
		eventType := nodeevents.Disqualified

		// Insert into node events
		event1, err := db.NodeEvents().Insert(ctx, testEmail1, testID1, eventType)
		require.NoError(t, err)

		event2, err := db.NodeEvents().Insert(ctx, testEmail1, testID2, eventType)
		require.NoError(t, err)

		// GetNextBatch should get them.
		events, err := db.NodeEvents().GetNextBatch(ctx, time.Now())
		require.NoError(t, err)

		var foundEvent1, foundEvent2 bool
		for _, ne := range events {
			switch ne.NodeID {
			case event1.NodeID:
				foundEvent1 = true
			case event2.NodeID:
				foundEvent2 = true
			default:
			}
		}
		require.True(t, foundEvent1)
		require.True(t, foundEvent2)

		// Update email sent
		require.NoError(t, db.NodeEvents().UpdateEmailSent(ctx, []uuid.UUID{event1.ID, event2.ID}, time.Now()))

		// They shouldn't be found since email_sent should have been updated.
		// It's an indirect way of checking. Not the best. We would need to add a new Read method
		// to get specific rows by ID.
		events, err = db.NodeEvents().GetNextBatch(ctx, time.Now())
		require.NoError(t, err)
		require.Len(t, events, 0)
	})
}

func TestNodeEventsGetNextBatch(t *testing.T) {
	satellitedbtest.Run(t, func(ctx *testcontext.Context, t *testing.T, db satellite.DB) {
		testID1 := teststorj.NodeIDFromString("test1")
		testID2 := teststorj.NodeIDFromString("test2")
		testEmail1 := "test1@storj.test"
		testEmail2 := "test2@storj.test"

		eventType := nodeevents.Disqualified

		event1, err := db.NodeEvents().Insert(ctx, testEmail1, testID1, eventType)
		require.NoError(t, err)

		// insert one event with same email and event type, but with different node ID. It should be selected.
		event2, err := db.NodeEvents().Insert(ctx, testEmail1, testID2, eventType)
		require.NoError(t, err)

		// insert one event with same email and event type, but email_sent is not null. Should not be selected.
		event3, err := db.NodeEvents().Insert(ctx, testEmail1, testID1, eventType)
		require.NoError(t, err)

		require.NoError(t, db.NodeEvents().UpdateEmailSent(ctx, []uuid.UUID{event3.ID}, time.Now()))

		// insert one event with same email, but different type. Should not be selected.
		_, err = db.NodeEvents().Insert(ctx, testEmail1, testID1, nodeevents.BelowMinVersion)
		require.NoError(t, err)

		// insert one event with same event type, but different email. Should not be selected.
		_, err = db.NodeEvents().Insert(ctx, testEmail2, testID1, eventType)
		require.NoError(t, err)

		batch, err := db.NodeEvents().GetNextBatch(ctx, time.Now())
		require.NoError(t, err)
		require.Len(t, batch, 2)

		var foundEvent1, foundEvent2 bool
		for _, ne := range batch {
			switch ne.NodeID {
			case event1.NodeID:
				foundEvent1 = true
			case event2.NodeID:
				foundEvent2 = true
			default:
			}
		}
		require.True(t, foundEvent1)
		require.True(t, foundEvent2)
	})
}
