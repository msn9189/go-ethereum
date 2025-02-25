package rawdb

import (
	"bytes"
	"math/big"

	"github.com/scroll-tech/go-ethereum/common"
	"github.com/scroll-tech/go-ethereum/ethdb"
	"github.com/scroll-tech/go-ethereum/log"
	"github.com/scroll-tech/go-ethereum/rlp"
)

// ChunkBlockRange represents the range of blocks within a chunk.
type ChunkBlockRange struct {
	StartBlockNumber uint64
	EndBlockNumber   uint64
}

// FinalizedBatchMeta holds metadata for finalized batches.
type FinalizedBatchMeta struct {
	BatchHash            common.Hash
	TotalL1MessagePopped uint64 // total number of L1 messages popped before and in this batch.
	StateRoot            common.Hash
	WithdrawRoot         common.Hash
}

// WriteRollupEventSyncedL1BlockNumber stores the latest synced L1 block number related to rollup events in the database.
func WriteRollupEventSyncedL1BlockNumber(db ethdb.KeyValueWriter, l1BlockNumber uint64) {
	value := big.NewInt(0).SetUint64(l1BlockNumber).Bytes()
	if err := db.Put(rollupEventSyncedL1BlockNumberKey, value); err != nil {
		log.Crit("failed to store rollup event synced L1 block number for rollup event", "L1 block number", l1BlockNumber, "value", value, "err", err)
	}
}

// ReadRollupEventSyncedL1BlockNumber fetches the highest synced L1 block number associated with rollup events from the database.
func ReadRollupEventSyncedL1BlockNumber(db ethdb.Reader) *uint64 {
	data, err := db.Get(rollupEventSyncedL1BlockNumberKey)
	if err != nil && isNotFoundErr(err) {
		return nil
	}
	if err != nil {
		log.Crit("failed to read rollup event synced L1 block number from database", "err", err)
	}

	number := new(big.Int).SetBytes(data)
	if !number.IsUint64() {
		log.Crit("unexpected rollup event synced L1 block number in database", "data", data, "number", number)
	}

	rollupEventSyncedL1BlockNumber := number.Uint64()
	return &rollupEventSyncedL1BlockNumber
}

// WriteBatchChunkRanges writes the block ranges for each chunk within a batch to the database.
// It serializes the chunk ranges using RLP and stores them under a key derived from the batch index.
func WriteBatchChunkRanges(db ethdb.KeyValueWriter, batchIndex uint64, chunkBlockRanges []*ChunkBlockRange) {
	value, err := rlp.EncodeToBytes(chunkBlockRanges)
	if err != nil {
		log.Crit("failed to RLP encode batch chunk ranges", "batch index", batchIndex, "err", err)
	}
	if err := db.Put(batchChunkRangesKey(batchIndex), value); err != nil {
		log.Crit("failed to store batch chunk ranges", "batch index", batchIndex, "value", value, "err", err)
	}
}

// DeleteBatchChunkRanges removes the block ranges of all chunks associated with a specific batch from the database.
// Note: Only non-finalized batches can be reverted.
func DeleteBatchChunkRanges(db ethdb.KeyValueWriter, batchIndex uint64) {
	if err := db.Delete(batchChunkRangesKey(batchIndex)); err != nil {
		log.Crit("failed to delete batch chunk ranges", "batch index", batchIndex, "err", err)
	}
}

// ReadBatchChunkRanges retrieves the block ranges of all chunks associated with a specific batch from the database.
// It returns a list of ChunkBlockRange pointers, or nil if no chunk ranges are found for the given batch index.
func ReadBatchChunkRanges(db ethdb.Reader, batchIndex uint64) []*ChunkBlockRange {
	data, err := db.Get(batchChunkRangesKey(batchIndex))
	if err != nil && isNotFoundErr(err) {
		return nil
	}
	if err != nil {
		log.Crit("failed to read batch chunk ranges from database", "err", err)
	}

	cr := new([]*ChunkBlockRange)
	if err := rlp.Decode(bytes.NewReader(data), cr); err != nil {
		log.Crit("Invalid ChunkBlockRange RLP", "batch index", batchIndex, "data", data, "err", err)
	}
	return *cr
}

// WriteFinalizedBatchMeta stores the metadata of a finalized batch in the database.
func WriteFinalizedBatchMeta(db ethdb.KeyValueWriter, batchIndex uint64, finalizedBatchMeta *FinalizedBatchMeta) {
	var err error
	value, err := rlp.EncodeToBytes(finalizedBatchMeta)
	if err != nil {
		log.Crit("failed to RLP encode batch metadata", "batch index", batchIndex, "finalized batch meta", finalizedBatchMeta, "err", err)
	}
	if err := db.Put(batchMetaKey(batchIndex), value); err != nil {
		log.Crit("failed to store batch metadata", "batch index", batchIndex, "value", value, "err", err)
	}
}

// ReadFinalizedBatchMeta fetches the metadata of a finalized batch from the database.
func ReadFinalizedBatchMeta(db ethdb.Reader, batchIndex uint64) *FinalizedBatchMeta {
	data, err := db.Get(batchMetaKey(batchIndex))
	if err != nil && isNotFoundErr(err) {
		return nil
	}
	if err != nil {
		log.Crit("failed to read finalized batch metadata from database", "batch index", batchIndex, "err", err)
	}

	fbm := new(FinalizedBatchMeta)
	if err := rlp.Decode(bytes.NewReader(data), fbm); err != nil {
		log.Crit("Invalid FinalizedBatchMeta RLP", "batch index", batchIndex, "data", data, "err", err)
	}
	return fbm
}

// WriteFinalizedL2BlockNumber stores the highest finalized L2 block number in the database.
func WriteFinalizedL2BlockNumber(db ethdb.KeyValueWriter, l2BlockNumber uint64) {
	value := big.NewInt(0).SetUint64(l2BlockNumber).Bytes()
	if err := db.Put(finalizedL2BlockNumberKey, value); err != nil {
		log.Crit("failed to store finalized L2 block number for rollup event", "L2 block number", l2BlockNumber, "value", value, "err", err)
	}
}

// ReadFinalizedL2BlockNumber fetches the highest finalized L2 block number from the database.
func ReadFinalizedL2BlockNumber(db ethdb.Reader) *uint64 {
	data, err := db.Get(finalizedL2BlockNumberKey)
	if err != nil && isNotFoundErr(err) {
		return nil
	}
	if err != nil {
		log.Crit("failed to read finalized L2 block number from database", "key", finalizedL2BlockNumberKey, "err", err)
	}

	number := new(big.Int).SetBytes(data)
	if !number.IsUint64() {
		log.Crit("unexpected finalized L2 block number in database", "data", data, "number", number)
	}

	finalizedL2BlockNumber := number.Uint64()
	return &finalizedL2BlockNumber
}

// WriteLastFinalizedBatchIndex stores the last finalized batch index in the database.
func WriteLastFinalizedBatchIndex(db ethdb.KeyValueWriter, lastFinalizedBatchIndex uint64) {
	value := big.NewInt(0).SetUint64(lastFinalizedBatchIndex).Bytes()
	if err := db.Put(lastFinalizedBatchIndexKey, value); err != nil {
		log.Crit("failed to store last finalized batch index for rollup event", "batch index", lastFinalizedBatchIndex, "value", value, "err", err)
	}
}

// ReadLastFinalizedBatchIndex fetches the last finalized batch index from the database.
func ReadLastFinalizedBatchIndex(db ethdb.Reader) *uint64 {
	data, err := db.Get(lastFinalizedBatchIndexKey)
	if err != nil && isNotFoundErr(err) {
		return nil
	}
	if err != nil {
		log.Crit("failed to read last finalized batch index from database", "key", lastFinalizedBatchIndexKey, "err", err)
	}

	number := new(big.Int).SetBytes(data)
	if !number.IsUint64() {
		log.Crit("unexpected finalized batch index in database", "data", data, "number", number)
	}

	lastFinalizedBatchIndex := number.Uint64()
	return &lastFinalizedBatchIndex
}
