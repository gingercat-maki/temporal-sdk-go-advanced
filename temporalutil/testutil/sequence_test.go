package testutil_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cretz/temporal-sdk-go-advanced/temporalutil/testutil"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func ExampleAddSequence() {
	var suite testsuite.WorkflowTestSuite
	suite.SetLogger(simpleLogger)
	env := suite.NewTestWorkflowEnvironment()

	// Add sequence of stuff
	testutil.AddSequence(env, func(seq testutil.Sequencer) {
		fmt.Println("Started")
		env.SignalWorkflow("signal", "signal1")
		env.SignalWorkflow("signal", "signal2")
		seq.Tick()
		env.SignalWorkflow("signal", "signal3")
		env.SignalWorkflow("signal", "signal4")
		seq.Sleep(10 * time.Hour)
		env.SignalWorkflow("signal", "signal5")
		env.SignalWorkflow("signal", "signal6")
		seq.Sleep(0)
		env.SignalWorkflow("signal", "finish")
		fmt.Println("Finished")
	})

	// Run workflow that captures signals
	env.ExecuteWorkflow(func(ctx workflow.Context) (signalsReceived []string, err error) {
		sig := workflow.GetSignalChannel(ctx, "signal")
		var sigVal string
		for sigVal != "finish" {
			sig.Receive(ctx, &sigVal)
			signalsReceived = append(signalsReceived, sigVal)
		}
		return
	})

	// Dump signals received
	if env.GetWorkflowError() != nil {
		panic(env.GetWorkflowError())
	}
	var signalsReceived []string
	env.GetWorkflowResult(&signalsReceived)
	fmt.Println("Signals Received: " + strings.Join(signalsReceived, ", "))
	// Output:
	// Started
	// DEBUG Auto fire timer TimerID 1 TimerDuration 1ns TimeSkipped 1ns
	// DEBUG Auto fire timer TimerID 2 TimerDuration 10h0m0s TimeSkipped 10h0m0s
	// Finished
	// Signals Received: signal1, signal2, signal3, signal4, signal5, signal6, finish
}

func TestAddSequence(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	// Add sequence of stuff
	testutil.AddSequence(env, func(seq testutil.Sequencer) {
		env.SignalWorkflow("signal", "signal1")
		seq.Tick()
		env.SignalWorkflow("signal", "signal2")
		seq.Sleep(10 * time.Hour)
		env.SignalWorkflow("signal", "finish")
	})

	// Run workflow that captures signals
	env.ExecuteWorkflow(func(ctx workflow.Context) (signalsReceived []string, err error) {
		sig := workflow.GetSignalChannel(ctx, "signal")
		var sigVal string
		for sigVal != "finish" {
			sig.Receive(ctx, &sigVal)
			signalsReceived = append(signalsReceived, sigVal)
		}
		return
	})

	// Check result
	require.NoError(t, env.GetWorkflowError())
	var signalsReceived []string
	env.GetWorkflowResult(&signalsReceived)
	require.Equal(t, []string{"signal1", "signal2", "finish"}, signalsReceived)
}

type simpleLoggerImpl struct{}

var simpleLogger log.Logger = simpleLoggerImpl{}

func simpleLog(level, msg string, keyvals ...interface{}) {
	fmt.Println(append([]interface{}{level, msg}, keyvals...)...)
}

func (simpleLoggerImpl) Debug(msg string, keyvals ...interface{}) {
	simpleLog("DEBUG", msg, keyvals...)
}

func (simpleLoggerImpl) Info(msg string, keyvals ...interface{}) {
	simpleLog("INFO ", msg, keyvals...)
}

func (simpleLoggerImpl) Warn(msg string, keyvals ...interface{}) {
	simpleLog("WARN ", msg, keyvals...)
}

func (simpleLoggerImpl) Error(msg string, keyvals ...interface{}) {
	simpleLog("ERROR", msg, keyvals...)
}
