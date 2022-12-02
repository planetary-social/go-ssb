package bolt_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestReceiveLog_GetMessage_ReturnsPredefinedErrorWhenNotFound(t *testing.T) {
	db := fixtures.Bolt(t)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		sequence1 := fixtures.SomeReceiveLogSequence()
		sequence2 := fixtures.SomeReceiveLogSequence()

		_, err = txadapters.ReceiveLog.GetMessage(sequence1)
		require.ErrorIs(t, err, common.ErrReceiveLogEntryNotFound)

		_, err = txadapters.ReceiveLog.GetMessage(sequence2)
		require.ErrorIs(t, err, common.ErrReceiveLogEntryNotFound)

		err = txadapters.ReceiveLog.PutUnderSpecificSequence(msg.Id(), sequence1)
		require.NoError(t, err)

		err = txadapters.MessageRepository.Put(msg)
		require.NoError(t, err)

		_, err = txadapters.ReceiveLog.GetMessage(sequence1)
		require.NoError(t, err)

		_, err = txadapters.ReceiveLog.GetMessage(sequence2)
		require.ErrorIs(t, err, common.ErrReceiveLogEntryNotFound)

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLog_GetSequences_ReturnsPredefinedErrorWhenNotFound(t *testing.T) {
	db := fixtures.Bolt(t)

	msg1 := fixtures.SomeRefMessage()
	msg2 := fixtures.SomeRefMessage()

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		_, err = txadapters.ReceiveLog.GetSequences(msg1)
		require.ErrorIs(t, err, common.ErrReceiveLogEntryNotFound)

		_, err = txadapters.ReceiveLog.GetSequences(msg2)
		require.ErrorIs(t, err, common.ErrReceiveLogEntryNotFound)

		err = txadapters.ReceiveLog.PutUnderSpecificSequence(msg1, fixtures.SomeReceiveLogSequence())
		require.NoError(t, err)

		_, err = txadapters.ReceiveLog.GetSequences(msg1)
		require.NoError(t, err)

		_, err = txadapters.ReceiveLog.GetSequences(msg2)
		require.ErrorIs(t, err, common.ErrReceiveLogEntryNotFound)

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLog_Get_ReturnsNoMessagesWhenEmpty(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		msgs, err := txadapters.ReceiveLog.List(common.MustNewReceiveLogSequence(0), 10)
		require.NoError(t, err)
		require.Empty(t, msgs)

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLog_Get_ReturnsErrorForInvalidLimit(t *testing.T) {
	db := fixtures.Bolt(t)

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		_, err = txadapters.ReceiveLog.List(common.MustNewReceiveLogSequence(0), 0)
		require.EqualError(t, err, "limit must be positive")

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLog_Put_InsertsCorrectMapping(t *testing.T) {
	db := fixtures.Bolt(t)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())
	expectedSequence := common.MustNewReceiveLogSequence(0)

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		if err := txadapters.ReceiveLog.Put(msg.Id()); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		if err := txadapters.MessageRepository.Put(msg); err != nil {
			return errors.Wrap(err, "could not put a message in message repository")
		}

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		seqs, err := txadapters.ReceiveLog.GetSequences(msg.Id())
		require.NoError(t, err)
		require.Equal(t, []common.ReceiveLogSequence{expectedSequence}, seqs)

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		_, err = txadapters.ReceiveLog.GetMessage(expectedSequence)
		require.NoError(t, err)
		// retrieved message won't have the same fields as the message we saved
		// as the raw data set in fixtures.SomeMessage is gibberish

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLog_Get_ReturnsMessagesObeyingLimitAndStartSeq(t *testing.T) {
	db := fixtures.Bolt(t)

	numMessages := 10

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		for i := 0; i < numMessages; i++ {
			msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

			if err := txadapters.ReceiveLog.Put(msg.Id()); err != nil {
				return errors.Wrap(err, "could not put a message in receive log")
			}

			if err := txadapters.MessageRepository.Put(msg); err != nil {
				return errors.Wrap(err, "could not put a message in message repository")
			}
		}

		return nil
	})
	require.NoError(t, err)

	t.Run("seq_0", func(t *testing.T) {
		err = db.Update(func(tx *bbolt.Tx) error {
			txadapters, err := di.BuildTxTestAdapters(tx)
			require.NoError(t, err)

			msgs, err := txadapters.ReceiveLog.List(common.MustNewReceiveLogSequence(0), 10)
			require.NoError(t, err)
			require.Len(t, msgs, 10)

			return nil
		})
		require.NoError(t, err)
	})

	t.Run("seq_5", func(t *testing.T) {
		err = db.Update(func(tx *bbolt.Tx) error {
			txadapters, err := di.BuildTxTestAdapters(tx)
			require.NoError(t, err)

			msgs, err := txadapters.ReceiveLog.List(common.MustNewReceiveLogSequence(5), 10)
			require.NoError(t, err)
			require.Len(t, msgs, 5)

			return nil
		})
		require.NoError(t, err)
	})
}

func TestReceiveLog_PutUnderSpecificSequence_InsertsCorrectMapping(t *testing.T) {
	db := fixtures.Bolt(t)

	msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())
	sequence := common.MustNewReceiveLogSequence(123)

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		if err := txadapters.ReceiveLog.PutUnderSpecificSequence(msg.Id(), sequence); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		if err := txadapters.MessageRepository.Put(msg); err != nil {
			return errors.Wrap(err, "could not put a message in message repository")
		}

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		seqs, err := txadapters.ReceiveLog.GetSequences(msg.Id())
		require.NoError(t, err)
		require.Equal(t, []common.ReceiveLogSequence{sequence}, seqs)

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		_, err = txadapters.ReceiveLog.GetMessage(sequence)
		require.NoError(t, err)
		// retrieved message won't have the same fields as the message we saved
		// as the raw data set in fixtures.SomeMessage is gibberish

		return nil
	})
	require.NoError(t, err)
}

func TestReceiveLogRepository_PutUnderSpecificSequenceAdvancesInternalSequenceCounter(t *testing.T) {
	db := fixtures.Bolt(t)

	msg1 := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())
	sequence := common.MustNewReceiveLogSequence(123)

	msg2 := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())

	err := db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		if err := txadapters.ReceiveLog.PutUnderSpecificSequence(msg1.Id(), sequence); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		if err := txadapters.MessageRepository.Put(msg1); err != nil {
			return errors.Wrap(err, "could not put a message in message repository")
		}

		return nil
	})
	require.NoError(t, err)

	err = db.View(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		msgs, err := txadapters.ReceiveLog.List(common.MustNewReceiveLogSequence(0), 100)
		require.NoError(t, err)

		require.Len(t, msgs, 1)
		require.Equal(t, sequence, msgs[0].Sequence)

		return nil
	})
	require.NoError(t, err)

	err = db.Update(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		if err := txadapters.ReceiveLog.Put(msg2.Id()); err != nil {
			return errors.Wrap(err, "could not put a message in receive log")
		}

		if err := txadapters.MessageRepository.Put(msg2); err != nil {
			return errors.Wrap(err, "could not put a message in message repository")
		}

		return nil
	})
	require.NoError(t, err)

	err = db.View(func(tx *bbolt.Tx) error {
		txadapters, err := di.BuildTxTestAdapters(tx)
		require.NoError(t, err)

		msgs, err := txadapters.ReceiveLog.List(common.MustNewReceiveLogSequence(0), 100)
		require.NoError(t, err)

		require.Len(t, msgs, 2)
		require.Equal(t, sequence, msgs[0].Sequence)
		require.Equal(t, common.MustNewReceiveLogSequence(sequence.Int()+1), msgs[1].Sequence)

		return nil
	})
	require.NoError(t, err)
}
