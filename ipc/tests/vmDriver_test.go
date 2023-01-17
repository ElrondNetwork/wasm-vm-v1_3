package tests

import (
	"testing"

	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/multiversx/mx-chain-vm-common-go/mock"
	"github.com/multiversx/mx-chain-vm-v1_3-go/vmhost"
	"github.com/multiversx/mx-chain-vm-v1_3-go/config"
	"github.com/multiversx/mx-chain-vm-v1_3-go/ipc/common"
	"github.com/multiversx/mx-chain-vm-v1_3-go/ipc/nodepart"
	contextmock "github.com/multiversx/mx-chain-vm-v1_3-go/mock/context"
	worldmock "github.com/multiversx/mx-chain-vm-v1_3-go/mock/world"
	"github.com/stretchr/testify/require"
)

var arwenVirtualMachine = []byte{5, 0}

func TestArwenDriver_DiagnoseWait(t *testing.T) {
	t.Skip("cannot unmarshal BuiltInFuncContainer")

	blockchain := &contextmock.BlockchainHookStub{}
	driver := newDriver(t, blockchain)

	err := driver.DiagnoseWait(100)
	require.Nil(t, err)
}

func TestArwenDriver_DiagnoseWaitWithTimeout(t *testing.T) {
	t.Skip("cannot unmarshal BuiltInFuncContainer")

	blockchain := &contextmock.BlockchainHookStub{}
	driver := newDriver(t, blockchain)

	err := driver.DiagnoseWait(5000)
	require.True(t, common.IsCriticalError(err))
	require.Contains(t, err.Error(), "timeout")
	require.True(t, driver.IsClosed())
}

func TestArwenDriver_RestartsIfStopped(t *testing.T) {
	t.Skip("cannot unmarshal BuiltInFuncContainer")

	logger.ToggleLoggerName(true)
	_ = logger.SetLogLevel("*:TRACE")

	blockchain := &contextmock.BlockchainHookStub{}
	driver := newDriver(t, blockchain)

	blockchain.GetUserAccountCalled = func(address []byte) (vmcommon.UserAccountHandler, error) {
		return &worldmock.Account{Code: bytecodeCounter}, nil
	}

	vmOutput, err := driver.RunSmartContractCreate(createDeployInput(bytecodeCounter))
	require.Nil(t, err)
	require.NotNil(t, vmOutput)
	vmOutput, err = driver.RunSmartContractCall(createCallInput("increment"))
	require.Nil(t, err)
	require.NotNil(t, vmOutput)

	require.False(t, driver.IsClosed())
	_ = driver.Close()
	require.True(t, driver.IsClosed())

	// Per this request, Arwen is restarted
	vmOutput, err = driver.RunSmartContractCreate(createDeployInput(bytecodeCounter))
	require.Nil(t, err)
	require.NotNil(t, vmOutput)
	require.False(t, driver.IsClosed())
}

func BenchmarkArwenDriver_RestartsIfStopped(b *testing.B) {
	blockchain := &contextmock.BlockchainHookStub{}
	driver := newDriver(b, blockchain)

	for i := 0; i < b.N; i++ {
		_ = driver.Close()
		require.True(b, driver.IsClosed())
		_ = driver.RestartArwenIfNecessary()
		require.False(b, driver.IsClosed())
	}
}

func BenchmarkArwenDriver_RestartArwenIfNecessary(b *testing.B) {
	blockchain := &contextmock.BlockchainHookStub{}
	driver := newDriver(b, blockchain)

	for i := 0; i < b.N; i++ {
		_ = driver.RestartArwenIfNecessary()
	}
}

func TestArwenDriver_GetVersion(t *testing.T) {
	t.Skip("cannot unmarshal BuiltInFuncContainer")
	// This test requires `make arwen` before running, or must be run directly
	// with `make test`
	blockchain := &contextmock.BlockchainHookStub{}
	driver := newDriver(t, blockchain)
	version := driver.GetVersion()
	require.NotZero(t, len(version))
	require.NotEqual(t, "undefined", version)
}

func newDriver(tb testing.TB, blockchain *contextmock.BlockchainHookStub) *nodepart.ArwenDriver {
	driver, err := nodepart.NewArwenDriver(
		blockchain,
		common.ArwenArguments{
			VMHostParameters: vmhost.VMHostParameters{
				VMType:               arwenVirtualMachine,
				BlockGasLimit:        uint64(10000000),
				GasSchedule:          config.MakeGasMapForTests(),
				ProtectedKeyPrefix:   []byte("E"+"L"+"R"+"O"+"N"+"D"),
				BuiltInFuncContainer: builtInFunctions.NewBuiltInFunctionContainer(),
				EnableEpochsHandler: &mock.EnableEpochsHandlerStub{
					IsSCDeployFlagEnabledField:            true,
					IsAheadOfTimeGasUsageFlagEnabledField: true,
					IsRepairCallbackFlagEnabledField:      true,
					IsBuiltInFunctionsFlagEnabledField:    true,
				},
			},
		},
		nodepart.Config{MaxLoopTime: 1000},
	)
	require.Nil(tb, err)
	require.NotNil(tb, driver)
	require.False(tb, driver.IsClosed())
	return driver
}
