// x/htlc/msg_claim.go
package htlc

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgClaimHTLC struct {
    Claimer sdk.AccAddress
    ID      string
    Secret  []byte
}

func NewMsgClaimHTLC(claimer sdk.AccAddress, id string, secret []byte) MsgClaimHTLC {
    return MsgClaimHTLC{
        Claimer: claimer,
        ID:      id,
        Secret:  secret,
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
