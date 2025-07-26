// x/htlc/keeper.go
package htlc

import (
    "bytes"
    "time"

    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/store/prefix"
)

type Keeper struct {
    storeKey sdk.StoreKey
    cdc      codec.BinaryCodec
}

type HTLC struct {
    ID          string         // unique identifier
    Sender      sdk.AccAddress // sender address
    Receiver    sdk.AccAddress // receiver address
    Amount      sdk.Coins      // locked tokens
    HashLock    []byte         // hash of the secret
    TimeLock    time.Time      // expiration time
    Claimed     bool           // whether claimed
    Refunded    bool           // whether refunded
}

func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey) Keeper {
    return Keeper{
        storeKey: storeKey,
        cdc:      cdc,
    }
}

// Store key prefix for HTLCs
var HTLCKeyPrefix = []byte{0x01}

func (k Keeper) getHTLCStore(ctx sdk.Context) prefix.Store {
    return prefix.NewStore(ctx.KVStore(k.storeKey), HTLCKeyPrefix)
}

func (k Keeper) CreateHTLC(ctx sdk.Context, msg MsgCreateHTLC) error {
    store := k.getHTLCStore(ctx)

    id := msg.Sender.String() + "-" + ctx.BlockTime().String()
    if store.Has([]byte(id)) {
        return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "HTLC already exists")
    }

    htlc := HTLC{
        ID:            id,
        Sender:        msg.Sender,
        Receiver:      msg.Receiver,
        Amount:        msg.Amount,
        HashLock:      msg.HashLock,
        TimeLock:      time.Unix(int64(msg.TimeLock), 0),
        Claimed:       false,
        Refunded:      false,
        ExternalChain: msg.ExternalChain,
        ExternalID:    msg.ExternalID,
    }

    // Securely lock tokens by sending from sender to module account
    if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, msg.Sender, "htlc", msg.Amount); err != nil {
        return err
    }

    // Additional handling for IBC tokens could be added here if needed

    bz, err := k.cdc.Marshal(&htlc)
    if err != nil {
        return err
    }

    store.Set([]byte(id), bz)
    return nil
}

func (k Keeper) ClaimHTLC(ctx sdk.Context, msg MsgClaimHTLC) error {
    store := k.getHTLCStore(ctx)
    bz := store.Get([]byte(msg.ID))
    if bz == nil {
        return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "HTLC not found")
    }

    var htlc HTLC
    if err := k.cdc.Unmarshal(bz, &htlc); err != nil {
        return err
    }

    if htlc.Claimed {
        return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "HTLC already claimed")
    }
    if htlc.Refunded {
        return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "HTLC already refunded")
    }

    // Verify secret with Merkle proof if MerkleRoot is set (partial fill)
    if len(htlc.MerkleRoot) > 0 {
        leaf := sdk.Sha256(msg.Secret)
        if !VerifyMerkleProof(leaf, msg.MerkleProof, htlc.MerkleRoot) {
            return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Invalid Merkle proof")
        }
        secretStr := string(msg.Secret)
        if htlc.UsedSecrets[secretStr] {
            return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Secret already used")
        }
        htlc.UsedSecrets[secretStr] = true
    } else {
        // Single secret verification
        if !bytes.Equal(htlc.HashLock, sdk.Sha256(msg.Secret)) {
            return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Invalid secret")
        }
    }

    if ctx.BlockTime().After(htlc.TimeLock) {
        return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "HTLC expired")
    }
    if !msg.Claimer.Equals(htlc.Receiver) {
        return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Not receiver")
    }

    // Transfer tokens from module account to receiver
    if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, "htlc", htlc.Receiver, htlc.Amount); err != nil {
        return err
    }

    // If all secrets used or single secret, mark claimed
    if len(htlc.MerkleRoot) == 0 || allSecretsUsed(htlc) {
        htlc.Claimed = true
    }

    bz, err := k.cdc.Marshal(&htlc)
    if err != nil {
        return err
    }
    store.Set([]byte(msg.ID), bz)
    return nil
}

// VerifyMerkleProof verifies a Merkle proof for a leaf and root
func VerifyMerkleProof(leaf []byte, proof [][]byte, root []byte) bool {
    computedHash := leaf
    for _, p := range proof {
        if bytes.Compare(computedHash, p) < 0 {
            computedHash = sdk.Sha256(append(computedHash, p...))
        } else {
            computedHash = sdk.Sha256(append(p, computedHash...))
        }
    }
    return bytes.Equal(computedHash, root)
}

// allSecretsUsed checks if all secrets have been used (placeholder)
func allSecretsUsed(htlc HTLC) bool {
    // In a real implementation, track total secrets count and compare with usedSecrets
    // For now, return false to allow multiple partial fills
    return false
}

func (k Keeper) RefundHTLC(ctx sdk.Context, msg MsgRefundHTLC) error {
    store := k.getHTLCStore(ctx)
    bz := store.Get([]byte(msg.ID))
    if bz == nil {
        return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "HTLC not found")
    }

    var htlc HTLC
    if err := k.cdc.Unmarshal(bz, &htlc); err != nil {
        return err
    }

    if htlc.Claimed {
        return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "HTLC already claimed")
    }
    if htlc.Refunded {
        return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "HTLC already refunded")
    }
    if ctx.BlockTime().Before(htlc.TimeLock) {
        return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "HTLC not expired")
    }
    if !msg.Sender.Equals(htlc.Sender) {
        return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Not sender")
    }

    // Transfer tokens from module account back to sender
    if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, "htlc", htlc.Sender, htlc.Amount); err != nil {
        return err
    }

    htlc.Refunded = true
    bz, err := k.cdc.Marshal(&htlc)
    if err != nil {
        return err
    }
    store.Set([]byte(msg.ID), bz)
    return nil
}
