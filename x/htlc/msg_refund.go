// x/htlc/msg_refund.go
package htlc

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

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
