package transfer_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypestm "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	transfer "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
)

// define constants used for testing
const (
	testChainID    = "test-chain-id"
	testClient     = "test-client"
	testClientType = clientexported.Tendermint

	testConnection = "testconnection"
	testPort1      = "bank"
	testPort2      = "testportid"
	testChannel1   = "firstchannel"
	testChannel2   = "secondchannel"

	testChannelOrder   = channel.UNORDERED
	testChannelVersion = "1.0"
)

// define variables used for testing
var (
	testAddr1 = sdk.AccAddress([]byte("testaddr1"))
	testAddr2 = sdk.AccAddress([]byte("testaddr2"))

	testCoins, _          = sdk.ParseCoins("100atom")
	testPrefixedCoins1, _ = sdk.ParseCoins(fmt.Sprintf("100%satom", types.GetDenomPrefix(testPort1, testChannel1)))
	testPrefixedCoins2, _ = sdk.ParseCoins(fmt.Sprintf("100%satom", types.GetDenomPrefix(testPort2, testChannel2)))
)

type HandlerTestSuite struct {
	suite.Suite

	cdc *codec.Codec
	ctx sdk.Context
	app *simapp.SimApp
}

func (suite *HandlerTestSuite) SetupTest() {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.Codec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{})
	suite.app = app

	suite.createClient()
	suite.createConnection(connection.OPEN)
}

func (suite *HandlerTestSuite) createClient() {
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
	suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{})

	consensusState := clienttypestm.ConsensusState{
		ChainID: testChainID,
		Height:  uint64(commitID.Version),
		Root:    commitment.NewRoot(commitID.Hash),
	}

	_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClient, testClientType, consensusState)
	suite.NoError(err)
}

func (suite *HandlerTestSuite) updateClient() {
	// always commit and begin a new block on updateClient
	suite.app.Commit()
	commitID := suite.app.LastCommitID()

	suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.app.LastBlockHeight() + 1}})
	suite.ctx = suite.app.BaseApp.NewContext(false, abci.Header{})

	state := clienttypestm.ConsensusState{
		ChainID: testChainID,
		Height:  uint64(commitID.Version),
		Root:    commitment.NewRoot(commitID.Hash),
	}

	suite.app.IBCKeeper.ClientKeeper.SetConsensusState(suite.ctx, testClient, state)
	suite.app.IBCKeeper.ClientKeeper.SetVerifiedRoot(suite.ctx, testClient, state.GetHeight(), state.GetRoot())
}

func (suite *HandlerTestSuite) createConnection(state connection.State) {
	connection := connection.ConnectionEnd{
		State:    state,
		ClientID: testClient,
		Counterparty: connection.Counterparty{
			ClientID:     testClient,
			ConnectionID: testConnection,
			Prefix:       suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix(),
		},
		Versions: connection.GetCompatibleVersions(),
	}

	suite.app.IBCKeeper.ConnectionKeeper.SetConnection(suite.ctx, testConnection, connection)
}

func (suite *HandlerTestSuite) createChannel(portID string, chanID string, connID string, counterpartyPort string, counterpartyChan string, state channel.State) {
	ch := channel.Channel{
		State:    state,
		Ordering: testChannelOrder,
		Counterparty: channel.Counterparty{
			PortID:    counterpartyPort,
			ChannelID: counterpartyChan,
		},
		ConnectionHops: []string{connID},
		Version:        testChannelVersion,
	}

	suite.app.IBCKeeper.ChannelKeeper.SetChannel(suite.ctx, portID, chanID, ch)
}

func (suite *HandlerTestSuite) queryProof(key string) (proof commitment.Proof, height int64) {
	res := suite.app.Query(abci.RequestQuery{
		Path:  fmt.Sprintf("store/%s/key", ibctypes.StoreKey),
		Data:  []byte(key),
		Prove: true,
	})

	height = res.Height
	proof = commitment.Proof{
		Proof: res.Proof,
	}

	return
}

func (suite *HandlerTestSuite) TestHandleMsgTransfer() {
	source := true

	handler := transfer.NewHandler(suite.app.TransferKeeper)

	msg := transfer.NewMsgTransfer(testPort1, testChannel1, testCoins, testAddr1, testAddr2, source)
	res := handler(suite.ctx, msg)
	suite.False(res.Code.IsOK(), "%+v", res) // channel does not exist

	suite.createChannel(testPort1, testChannel1, testConnection, testPort2, testChannel2, channel.OPEN)
	res = handler(suite.ctx, msg)
	suite.False(res.Code.IsOK(), "%+v", res) // next send sequence not found

	nextSeqSend := uint64(1)
	suite.app.IBCKeeper.ChannelKeeper.SetNextSequenceSend(suite.ctx, testPort1, testChannel1, nextSeqSend)
	res = handler(suite.ctx, msg)
	suite.False(res.Code.IsOK(), "%+v", res) // sender has insufficient coins

	_ = suite.app.BankKeeper.SetCoins(suite.ctx, testAddr1, testCoins)
	res = handler(suite.ctx, msg)
	suite.True(res.Code.IsOK(), "%+v", res) // successfully executed

	// test when the source is false
	source = false

	msg = transfer.NewMsgTransfer(testPort1, testChannel1, testPrefixedCoins2, testAddr1, testAddr2, source)
	_ = suite.app.BankKeeper.SetCoins(suite.ctx, testAddr1, testPrefixedCoins2)

	res = handler(suite.ctx, msg)
	suite.False(res.Code.IsOK(), "%+v", res) // incorrect denom prefix

	msg = transfer.NewMsgTransfer(testPort1, testChannel1, testPrefixedCoins1, testAddr1, testAddr2, source)
	suite.app.SupplyKeeper.SetSupply(suite.ctx, supply.NewSupply(testPrefixedCoins1))
	_ = suite.app.BankKeeper.SetCoins(suite.ctx, testAddr1, testPrefixedCoins1)
	res = handler(suite.ctx, msg)
	suite.True(res.Code.IsOK(), "%+v", res) // successfully executed
}

// XXX: fix before merge
/*
func (suite *HandlerTestSuite) TestHandleRecvPacket() {
	packetSeq := uint64(1)
	packetTimeout := uint64(100)
	handler := transfer.NewHandler(suite.app.TransferKeeper)

	// test when the source is true
	source := true

	packetData := types.NewPacketDataTransfer(testPrefixedCoins2, testAddr1, testAddr2, source, packetTimeout)
	packet := channel.NewPacket(packetData, packetSeq, testPort2, testChannel2, testPort1, testChannel1)
	packetCommitmentPath := channel.PacketCommitmentPath(testPort2, testChannel2, packetSeq)

	suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort2, testChannel2, packetSeq, packetData.GetCommitment())
	suite.updateClient()
	proofPacket, proofHeight := suite.queryProof(packetCommitmentPath)

	msg := channel.NewMsgPacket(packet, proofPacket, uint64(proofHeight), testAddr1)
	res := handler(suite.ctx, msg)
	suite.False(res.Code.IsOK(), "%+v", res) // invalid denom prefix

	packetData = types.NewPacketDataTransfer(testPrefixedCoins1, testAddr1, testAddr2, source, packetTimeout)
	packet = channel.NewPacket(packetData, packetSeq, testPort2, testChannel2, testPort1, testChannel1)

	suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort2, testChannel2, packetSeq, packetData.GetCommitment())
	suite.updateClient()
	proofPacket, proofHeight = suite.queryProof(packetCommitmentPath)

	msg = channel.NewMsgPacket(packet, proofPacket, uint64(proofHeight), testAddr1)
	res = handler(suite.ctx, msg)
	suite.True(res.Code.IsOK(), "%+v", res) // successfully executed

	// test when the source is false
	source = false

	packetData = types.NewPacketDataTransfer(testPrefixedCoins1, testAddr1, testAddr2, source, packetTimeout)
	packet = channel.NewPacket(packetData, packetSeq, testPort2, testChannel2, testPort1, testChannel1)

	suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort2, testChannel2, packetSeq, packetData.GetCommitment())
	suite.updateClient()
	proofPacket, proofHeight = suite.queryProof(packetCommitmentPath)

	msg = channel.NewMsgPacket(packet, proofPacket, uint64(proofHeight), testAddr1)
	res = handler(suite.ctx, msg)
	suite.False(res.Code.IsOK(), "%+v", res) // invalid denom prefix

	packetData = types.NewPacketDataTransfer(testPrefixedCoins2, testAddr1, testAddr2, source, packetTimeout)
	packet = channel.NewPacket(packetData, packetSeq, testPort2, testChannel2, testPort1, testChannel1)

	suite.app.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.ctx, testPort2, testChannel2, packetSeq, packetData.GetCommitment())
	suite.updateClient()
	proofPacket, proofHeight = suite.queryProof(packetCommitmentPath)

	msg = channel.NewMsgPacket(packet, proofPacket, uint64(proofHeight), testAddr1)
	res = handler(suite.ctx, msg)
	suite.False(res.Code.IsOK(), "%+v", res) // insufficient coins in the corresponding escrow account

	escrowAddress := types.GetEscrowAddress(testPort1, testChannel1)
	_ = suite.app.BankKeeper.SetCoins(suite.ctx, escrowAddress, testCoins)
	res = handler(suite.ctx, msg)
	suite.True(res.Code.IsOK(), "%+v", res) // successfully executed
}
*/
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}