// x/htlc/handler.go
package htlc

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewHandler(k Keeper) sdk.Handler {
    return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
        switch msg := msg.(type) {
        case MsgCreateHTLC:
            return handleMsgCreateHTLC(ctx, k, msg)
        case MsgClaimHTLC:
            return handleMsgClaimHTLC(ctx, k, msg)
        case MsgRefundHTLC:
            return handleMsgRefundHTLC(ctx, k, msg)
        default:
            return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized htlc message type: %T", msg)
        }
    }
}

func handleMsgCreateHTLC(ctx sdk.Context, k Keeper, msg MsgCreateHTLC) (*sdk.Result, error) {
    err := k.CreateHTLC(ctx, msg)
    if err != nil {
        return nil, err
    }
    return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgClaimHTLC(ctx sdk.Context, k Keeper, msg MsgClaimHTLC) (*sdk.Result, error) {
    err := k.ClaimHTLC(ctx, msg)
    if err != nil {
        return nil, err
    }
    return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgRefundHTLC(ctx sdk.Context, k Keeper, msg MsgRefundHTLC) (*sdk.Result, error) {
    err := k.RefundHTLC(ctx, msg)
    if err != nil {
        return nil, err
    }
    return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
