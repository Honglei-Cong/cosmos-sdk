package gov

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestTickExpiredDepositPeriod(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	govHandler := NewHandler(keeper)

	inactiveQueue := keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin("steak", 5)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetDepositProcedure(ctx).MaxDepositPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	EndBlocker(ctx, keeper)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
}

func TestTickMultipleExpiredDepositPeriod(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	govHandler := NewHandler(keeper)

	inactiveQueue := keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin("steak", 5)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(2) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg2 := NewMsgSubmitProposal("Test2", "test2", ProposalTypeText, addrs[1], sdk.Coins{sdk.NewInt64Coin("steak", 5)})
	res = govHandler(ctx, newProposalMsg2)
	require.True(t, res.IsOK())

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetDepositProcedure(ctx).MaxDepositPeriod).Add(time.Duration(-1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	EndBlocker(ctx, keeper)
	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(5) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	EndBlocker(ctx, keeper)
	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
}

func TestTickPassedDepositPeriod(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	govHandler := NewHandler(keeper)

	inactiveQueue := keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	activeQueue := keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin("steak", 5)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	var proposalID int64
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(res.Data, &proposalID)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newDepositMsg := NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.NewInt64Coin("steak", 5)})
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	activeQueue = keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()
}

func TestTickPassedVotingPeriod(t *testing.T) {
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 10)
	SortAddresses(addrs)
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	govHandler := NewHandler(keeper)

	inactiveQueue := keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	activeQueue := keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	newProposalMsg := NewMsgSubmitProposal("Test", "test", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin("steak", 5)})

	res := govHandler(ctx, newProposalMsg)
	require.True(t, res.IsOK())
	var proposalID int64
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(res.Data, &proposalID)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	newDepositMsg := NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.NewInt64Coin("steak", 5)})
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetDepositProcedure(ctx).MaxDepositPeriod).Add(keeper.GetVotingProcedure(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	activeQueue = keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, activeQueue.Valid())
	var activeProposalID int64
	keeper.cdc.UnmarshalBinaryLengthPrefixed(activeQueue.Value(), &activeProposalID)
	require.Equal(t, StatusVotingPeriod, keeper.GetProposal(ctx, activeProposalID).GetStatus())
	depositsIterator := keeper.GetDeposits(ctx, proposalID)
	require.True(t, depositsIterator.Valid())
	depositsIterator.Close()
	activeQueue.Close()

	EndBlocker(ctx, keeper)

	activeQueue = keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()
}
