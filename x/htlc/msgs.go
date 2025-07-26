// x/htlc/msgs.go
package htlc

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgCreateHTLC struct {
    Sender        sdk.AccAddress
    Receiver      sdk.AccAddress
    Amount        sdk.Coins
    HashLock      []byte
    TimeLock      uint64
    ExternalChain string // e.g., "ethereum"
    ExternalID    string // ID of corresponding HTLC on external chain
}

func NewMsgCreateHTLC(sender, receiver sdk.AccAddress, amount sdk.Coins, hashLock []byte, timeLock uint64, externalChain, externalID string) MsgCreateHTLC {
    return MsgCreateHTLC{
        Sender:        sender,
        Receiver:      receiver,
        Amount:        amount,
        HashLock:      hashLock,
        TimeLock:      timeLock,
        ExternalChain: externalChain,
        ExternalID:    externalID,
    }
}

func (msg MsgCreateHTLC) Route() string { return "htlc" }

func (msg MsgCreateHTLC) Type() string { return "create_htlc" }

func (msg MsgCreateHTLC) ValidateBasic() error {
    if msg.Sender.Empty() {
        return sdk.ErrInvalidAddress("missing sender address")
    }
    if msg.Receiver.Empty() {
        return sdk.ErrInvalidAddress("missing receiver address")
    }
    if !msg.Amount.IsAllPositive() {
        return sdk.ErrInsufficientFunds("amount must be positive")
    }
    return nil
}

func (msg MsgCreateHTLC) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{msg.Sender}
}

type MsgClaimHTLC struct {
    Claimer     sdk.AccAddress
    ID          string
    Secret      []byte
    MerkleProof [][]byte
}

func NewMsgClaimHTLC(claimer sdk.AccAddress, id string, secret []byte, merkleProof [][]byte) MsgClaimHTLC {
    return MsgClaimHTLC{
        Claimer:    claimer,
        ID:         id,
        Secret:     secret,
        MerkleProof: merkleProof,
    }
}

func (msg MsgClaimHTLC) Route() string { return "htlc" }

func (msg MsgClaimHTLC) Type() string { return "claim_htlc" }

func (msg MsgClaimHTLC) ValidateBasic() error {
    if msg.Claimer.Empty() {
        return sdk.ErrInvalidAddress("missing claimer address")
    }
    if len(msg.ID) == 0 {
        return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "missing HTLC ID")
    }
    if len(msg.Secret) == 0 {
        return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "missing secret")
    }
    return nil
}

func (msg MsgClaimHTLC) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{msg.Claimer}
}

type MsgRefundHTLC struct {
    Sender sdk.AccAddress
    ID     string
}

func NewMsgRefundHTLC(sender sdk.AccAddress, id string) MsgRefundHTLC {
    return MsgRefundHTLC{
        Sender: sender,
        ID:     id,
    }
}

func (msg MsgRefundHTLC) Route() string { return "htlc" }

func (msg MsgRefundHTLC) Type() string { return "refund_htlc" }

func (msg MsgRefundHTLC) ValidateBasic() error {
    if msg.Sender.Empty() {
        return sdk.ErrInvalidAddress("missing sender address")
    }
    if len(msg.ID) == 0 {
        return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "missing HTLC ID")
    }
    return nil
}

func (msg MsgRefundHTLC) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{msg.Sender}
}
