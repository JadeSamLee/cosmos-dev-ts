// x/htlc/keeper_test.go
package htlc_test

import (
    "testing"
    "time"

    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/store"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/types/module"
    "github.com/cosmos/cosmos-sdk/x/bank/keeper"
    banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
    "github.com/stretchr/testify/require"
    "github.com/tendermint/tendermint/libs/log"
    tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
    dbm "github.com/tendermint/tm-db"

    "github.com/your_repo/x/htlc"
)

func createTestInput(t *testing.T) (sdk.Context, htlc.Keeper, keeper.Keeper) {
    db := dbm.NewMemDB()
    cms := store.NewCommitMultiStore(db)
    key := sdk.NewKVStoreKey("htlc")
    bankKey := sdk.NewKVStoreKey(banktypes.StoreKey)
    cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
    cms.MountStoreWithDB(bankKey, sdk.StoreTypeIAVL, db)
    err := cms.LoadLatestVersion()
    require.NoError(t, err)

    cdc := codec.NewProtoCodec(module.NewBasicManager())

    bankKeeper := keeper.NewBaseKeeper(cdc, bankKey, nil, nil)
    k := htlc.NewKeeper(cdc, key)
    ctx := sdk.NewContext(cms, tmproto.Header{Time: time.Now()}, false, log.NewNopLogger())

    // Fund sender account
    sender := sdk.AccAddress([]byte("sender____________"))
    coins := sdk.NewCoins(sdk.NewInt64Coin("atom", 1000))
    err = bankKeeper.MintCoins(ctx, banktypes.ModuleName, coins)
    require.NoError(t, err)
    err = bankKeeper.SendCoinsFromModuleToAccount(ctx, banktypes.ModuleName, sender, coins)
    require.NoError(t, err)

    return ctx, k, bankKeeper
}

func TestCreateHTLC(t *testing.T) {
    ctx, k, _ := createTestInput(t)
    sender := sdk.AccAddress([]byte("sender____________"))
    receiver := sdk.AccAddress([]byte("receiver__________"))
    amount := sdk.NewCoins(sdk.NewInt64Coin("atom", 100))
    hashLock := sdk.Sha256([]byte("secret"))
    timeLock := uint64(ctx.BlockTime().Add(time.Hour).Unix())

    msg := htlc.MsgCreateHTLC{
        Sender:   sender,
        Receiver: receiver,
        Amount:   amount,
        HashLock: hashLock,
        TimeLock: timeLock,
    }

    err := k.CreateHTLC(ctx, msg)
    require.NoError(t, err)

    store := k.GetStore(ctx)
    id := sender.String() + "-" + ctx.BlockTime().String()
    bz := store.Get([]byte(id))
    require.NotNil(t, bz)
}

func TestClaimHTLC_Success(t *testing.T) {
    ctx, k, _ := createTestInput(t)
    sender := sdk.AccAddress([]byte("sender____________"))
    receiver := sdk.AccAddress([]byte("receiver__________"))
    secret := []byte("secret")
    hashLock := sdk.Sha256(secret)
    amount := sdk.NewCoins(sdk.NewInt64Coin("atom", 100))
    timeLock := uint64(ctx.BlockTime().Add(time.Hour).Unix())

    createMsg := htlc.MsgCreateHTLC{
        Sender:   sender,
        Receiver: receiver,
        Amount:   amount,
        HashLock: hashLock,
        TimeLock: timeLock,
    }
    err := k.CreateHTLC(ctx, createMsg)
    require.NoError(t, err)

    id := sender.String() + "-" + ctx.BlockTime().String()
    claimMsg := htlc.MsgClaimHTLC{
        Claimer: receiver,
        ID:      id,
        Secret:  secret,
    }
    err = k.ClaimHTLC(ctx, claimMsg)
    require.NoError(t, err)

    store := k.GetStore(ctx)
    bz := store.Get([]byte(id))
    var htlcObj htlc.HTLC
    err = k.Cdc().Unmarshal(bz, &htlcObj)
    require.NoError(t, err)
    require.True(t, htlcObj.Claimed)
}

func TestClaimHTLC_InvalidSecret(t *testing.T) {
    ctx, k, _ := createTestInput(t)
    sender := sdk.AccAddress([]byte("sender____________"))
    receiver := sdk.AccAddress([]byte("receiver__________"))
    secret := []byte("secret")
    wrongSecret := []byte("wrongsecret")
    hashLock := sdk.Sha256(secret)
    amount := sdk.NewCoins(sdk.NewInt64Coin("atom", 100))
    timeLock := uint64(ctx.BlockTime().Add(time.Hour).Unix())

    createMsg := htlc.MsgCreateHTLC{
        Sender:   sender,
        Receiver: receiver,
        Amount:   amount,
        HashLock: hashLock,
        TimeLock: timeLock,
    }
    err := k.CreateHTLC(ctx, createMsg)
    require.NoError(t, err)

    id := sender.String() + "-" + ctx.BlockTime().String()
    claimMsg := htlc.MsgClaimHTLC{
        Claimer: receiver,
        ID:      id,
        Secret:  wrongSecret,
    }
    err = k.ClaimHTLC(ctx, claimMsg)
    require.Error(t, err)
}

func TestRefundHTLC_Success(t *testing.T) {
    ctx, k, _ := createTestInput(t)
    sender := sdk.AccAddress([]byte("sender____________"))
    receiver := sdk.AccAddress([]byte("receiver__________"))
    secret := []byte("secret")
    hashLock := sdk.Sha256(secret)
    amount := sdk.NewCoins(sdk.NewInt64Coin("atom", 100))
    timeLock := uint64(ctx.BlockTime().Add(-time.Hour).Unix()) // expired

    createMsg := htlc.MsgCreateHTLC{
        Sender:   sender,
        Receiver: receiver,
        Amount:   amount,
        HashLock: hashLock,
        TimeLock: timeLock,
    }
    err := k.CreateHTLC(ctx, createMsg)
    require.NoError(t, err)

    id := sender.String() + "-" + ctx.BlockTime().String()
    refundMsg := htlc.MsgRefundHTLC{
        Sender: sender,
        ID:     id,
    }
    err = k.RefundHTLC(ctx, refundMsg)
    require.NoError(t, err)

    store := k.GetStore(ctx)
    bz := store.Get([]byte(id))
    var htlcObj htlc.HTLC
    err = k.Cdc().Unmarshal(bz, &htlcObj)
    require.NoError(t, err)
    require.True(t, htlcObj.Refunded)
}

func TestRefundHTLC_NotExpired(t *testing.T) {
    ctx, k, _ := createTestInput(t)
    sender := sdk.AccAddress([]byte("sender____________"))
    receiver := sdk.AccAddress([]byte("receiver__________"))
    secret := []byte("secret")
    hashLock := sdk.Sha256(secret)
    amount := sdk.NewCoins(sdk.NewInt64Coin("atom", 100))
    timeLock := uint64(ctx.BlockTime().Add(time.Hour).Unix()) // not expired

    createMsg := htlc.MsgCreateHTLC{
        Sender:   sender,
        Receiver: receiver,
        Amount:   amount,
        HashLock: hashLock,
        TimeLock: timeLock,
    }
    err := k.CreateHTLC(ctx, createMsg)
    require.NoError(t, err)

    id := sender.String() + "-" + ctx.BlockTime().String()
    refundMsg := htlc.MsgRefundHTLC{
        Sender: sender,
        ID:     id,
    }
    err = k.RefundHTLC(ctx, refundMsg)
    require.Error(t, err)
}
