// x/htlc/types.go
package htlc

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "time"
)

type HTLC struct {
    ID          string         // unique identifier
    Sender      sdk.AccAddress // sender address
    Receiver    sdk.AccAddress // receiver address
    Amount      sdk.Coins      // locked tokens
    HashLock    []byte         // hash of the secret
    TimeLock    time.Time      // expiration time
    Claimed     bool           // whether claimed
    Refunded    bool           // whether refunded

    // New fields for cross-chain swaps
    ExternalChain string // e.g., "ethereum"
    ExternalID    string // ID of corresponding HTLC on external chain

    // New fields for partial fills
    MerkleRoot   []byte           // Merkle root of secrets for partial fills
    UsedSecrets  map[string]bool  // Track used secrets (stringified)
}
