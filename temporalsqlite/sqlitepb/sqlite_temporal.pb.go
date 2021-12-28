// Code generated by protoc-gen-go_temporal. DO NOT EDIT.

package sqlitepb

import (
	context "context"
	fmt "fmt"
	client "go.temporal.io/sdk/client"
	worker "go.temporal.io/sdk/worker"
	workflow "go.temporal.io/sdk/workflow"
	reflect "reflect"
)

// Constants used as workflow, activity, query, and signal names.
const (
	SqliteName       = "temporal.sdk.sqlite.Sqlite.Sqlite"
	QueryName        = "temporal.sdk.sqlite.Sqlite.Query"
	SerializeName    = "temporal.sdk.sqlite.Sqlite.Serialize"
	UpdateName       = "temporal.sdk.sqlite.Sqlite.Update"
	ExecSignalName   = "temporal.sdk.sqlite.Sqlite.Exec"
	ExecResponseName = ExecSignalName + "-response"
)

type Client interface {
	ExecuteSqlite(ctx context.Context, opts *client.StartWorkflowOptions, req *SqliteOptions) (SqliteRun, error)

	// GetSqlite returns an existing run started by ExecuteSqlite.
	GetSqlite(ctx context.Context, workflowID, runID string) (SqliteRun, error)

	Query(ctx context.Context, workflowID, runID string, req *QueryRequest) (*QueryResponse, error)

	Serialize(ctx context.Context, workflowID, runID string) (*SerializeResponse, error)

	// Errors fail the workflow by default
	Update(ctx context.Context, workflowID, runID string, req *UpdateRequest) error

	Exec(ctx context.Context, workflowID, runID string, req *ExecRequest) (*ExecResponse, error)
}

// ClientOptions are used for NewClient.
type ClientOptions struct {
	// Required client.
	Client client.Client
	// Handler that must be present for client calls to succeed.
	CallResponseHandler CallResponseHandler
}

// CallResponseHandler handles activity responses.
type CallResponseHandler interface {
	// TaskQueue returns the task queue for response activities.
	TaskQueue() string

	// PrepareCall creates a new ID and channels to receive response/error.
	// Each channel only has a buffer of one and are never closed and only one is ever sent to.
	// If context is closed, the context error is returned on error channel.
	PrepareCall(ctx context.Context) (id string, chOk <-chan interface{}, chErr <-chan error)

	// AddResponseType adds an activity for the given type and ID field.
	// Does not error if activity name already exists for the same params.
	AddResponseType(activityName string, typ reflect.Type, idField string) error
}

type clientImpl struct {
	client              client.Client
	callResponseHandler CallResponseHandler
}

// NewClient creates a new Client.
func NewClient(opts ClientOptions) Client {
	if opts.Client == nil {
		panic("missing client")
	}
	c := &clientImpl{client: opts.Client, callResponseHandler: opts.CallResponseHandler}
	if opts.CallResponseHandler != nil {
		if err := opts.CallResponseHandler.AddResponseType(ExecResponseName, reflect.TypeOf((*ExecResponse)(nil)), "Id"); err != nil {
			panic(err)
		}
	}
	return c
}

func (c *clientImpl) ExecuteSqlite(ctx context.Context, opts *client.StartWorkflowOptions, req *SqliteOptions) (SqliteRun, error) {
	if opts == nil {
		opts = &client.StartWorkflowOptions{}
	}
	run, err := c.client.ExecuteWorkflow(ctx, *opts, SqliteName, req)
	if run == nil || err != nil {
		return nil, err
	}
	return &sqliteRun{c, run}, nil
}

func (c *clientImpl) GetSqlite(ctx context.Context, workflowID, runID string) (SqliteRun, error) {
	return &sqliteRun{c, c.client.GetWorkflow(ctx, workflowID, runID)}, nil
}

func (c *clientImpl) Query(ctx context.Context, workflowID, runID string, req *QueryRequest) (*QueryResponse, error) {
	var resp QueryResponse
	if val, err := c.client.QueryWorkflow(ctx, workflowID, runID, QueryName, req); err != nil {
		return nil, err
	} else if err = val.Get(&resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *clientImpl) Serialize(ctx context.Context, workflowID, runID string) (*SerializeResponse, error) {
	var resp SerializeResponse
	if val, err := c.client.QueryWorkflow(ctx, workflowID, runID, SerializeName); err != nil {
		return nil, err
	} else if err = val.Get(&resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *clientImpl) Update(ctx context.Context, workflowID, runID string, req *UpdateRequest) error {
	return c.client.SignalWorkflow(ctx, workflowID, runID, UpdateName, req)
}

func (c *clientImpl) Exec(ctx context.Context, workflowID, runID string, req *ExecRequest) (*ExecResponse, error) {
	if c.callResponseHandler == nil {
		return nil, fmt.Errorf("missing response handler")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	id, chOk, chErr := c.callResponseHandler.PrepareCall(ctx)
	req.Id = id
	req.ResponseTaskQueue = c.callResponseHandler.TaskQueue()
	if err := c.client.SignalWorkflow(ctx, workflowID, runID, ExecSignalName, req); err != nil {
		return nil, err
	}
	select {
	case resp := <-chOk:
		return resp.(*ExecResponse), nil
	case err := <-chErr:
		return nil, err
	}
}

// SqliteRun represents an execution of Sqlite.
type SqliteRun interface {
	// ID is the workflow ID.
	ID() string

	// RunID is the workflow run ID.
	RunID() string

	// Get returns the completed workflow value, waiting if necessary.
	Get(ctx context.Context) error

	Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error)

	Serialize(ctx context.Context) (*SerializeResponse, error)

	// Errors fail the workflow by default
	Update(ctx context.Context, req *UpdateRequest) error

	Exec(ctx context.Context, req *ExecRequest) (*ExecResponse, error)
}

type sqliteRun struct {
	client *clientImpl
	run    client.WorkflowRun
}

func (r *sqliteRun) ID() string { return r.run.GetID() }

func (r *sqliteRun) RunID() string { return r.run.GetRunID() }

func (r *sqliteRun) Get(ctx context.Context) error {
	return r.run.Get(ctx, nil)
}

func (r *sqliteRun) Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	return r.client.Query(ctx, r.ID(), "", req)
}

func (r *sqliteRun) Serialize(ctx context.Context) (*SerializeResponse, error) {
	return r.client.Serialize(ctx, r.ID(), "")
}

func (r *sqliteRun) Update(ctx context.Context, req *UpdateRequest) error {
	return r.client.Update(ctx, r.ID(), "", req)
}

func (r *sqliteRun) Exec(ctx context.Context, req *ExecRequest) (*ExecResponse, error) {
	return r.client.Exec(ctx, r.ID(), "", req)
}

type SqliteImpl interface {
	Run(workflow.Context) error

	Query(*QueryRequest) (*QueryResponse, error)

	Serialize() (*SerializeResponse, error)
}

// SqliteInput is input provided to SqliteImpl.Run.
type SqliteInput struct {
	Req    *SqliteOptions
	Update Update
	Exec   Exec
}

type sqliteWorker struct {
	newImpl func(workflow.Context, *SqliteInput) (SqliteImpl, error)
}

func (w sqliteWorker) Sqlite(ctx workflow.Context, req *SqliteOptions) error {
	in := &SqliteInput{Req: req}
	in.Update.Channel = workflow.GetSignalChannel(ctx, UpdateName)
	in.Exec.Channel = workflow.GetSignalChannel(ctx, ExecSignalName)
	impl, err := w.newImpl(ctx, in)
	if err != nil {
		return err
	}
	if err := workflow.SetQueryHandler(ctx, QueryName, impl.Query); err != nil {
		return err
	}
	if err := workflow.SetQueryHandler(ctx, SerializeName, impl.Serialize); err != nil {
		return err
	}
	return impl.Run(ctx)
}

// BuildSqlite returns a function for the given impl.
func BuildSqlite(newImpl func(workflow.Context, *SqliteInput) (SqliteImpl, error)) func(ctx workflow.Context, req *SqliteOptions) error {
	return sqliteWorker{newImpl}.Sqlite
}

// RegisterSqlite registers a workflow with the given impl.
func RegisterSqlite(r worker.WorkflowRegistry, newImpl func(workflow.Context, *SqliteInput) (SqliteImpl, error)) {
	r.RegisterWorkflowWithOptions(BuildSqlite(newImpl), workflow.RegisterOptions{Name: SqliteName})
}

// Errors fail the workflow by default
type Update struct{ Channel workflow.ReceiveChannel }

// Receive blocks until signal is received.
func (s Update) Receive(ctx workflow.Context) *UpdateRequest {
	var resp UpdateRequest
	s.Channel.Receive(ctx, &resp)
	return &resp
}

// ReceiveAsync returns received signal or nil if none.
func (s Update) ReceiveAsync() *UpdateRequest {
	var resp UpdateRequest
	if !s.Channel.ReceiveAsync(&resp) {
		return nil
	}
	return &resp
}

// Select adds the callback to the selector to be invoked when signal received. Callback can be nil.
func (s Update) Select(sel workflow.Selector, fn func(*UpdateRequest)) workflow.Selector {
	return sel.AddReceive(s.Channel, func(workflow.ReceiveChannel, bool) {
		req := s.ReceiveAsync()
		if fn != nil {
			fn(req)
		}
	})
}

type Exec struct{ Channel workflow.ReceiveChannel }

// Receive blocks until call is received.
func (s Exec) Receive(ctx workflow.Context) *ExecRequest {
	var resp ExecRequest
	s.Channel.Receive(ctx, &resp)
	return &resp
}

// ReceiveAsync returns received signal or nil if none.
func (s Exec) ReceiveAsync() *ExecRequest {
	var resp ExecRequest
	if !s.Channel.ReceiveAsync(&resp) {
		return nil
	}
	return &resp
}

// Select adds the callback to the selector to be invoked when signal received. Callback can be nil
func (s Exec) Select(sel workflow.Selector, fn func(*ExecRequest)) workflow.Selector {
	return sel.AddReceive(s.Channel, func(workflow.ReceiveChannel, bool) {
		req := s.ReceiveAsync()
		if fn != nil {
			fn(req)
		}
	})
}

// Respond sends a response. Activity options not used if request received via
// another workflow. If activity options needed and not present, they are taken
// from the context.
func (s Exec) Respond(ctx workflow.Context, opts *workflow.ActivityOptions, req *ExecRequest, resp *ExecResponse) workflow.Future {
	resp.Id = req.Id
	if req.ResponseWorkflowId != "" {
		return workflow.SignalExternalWorkflow(ctx, req.ResponseWorkflowId, "", ExecResponseName+"-"+req.Id, resp)
	}
	newOpts := workflow.GetActivityOptions(ctx)
	if opts != nil {
		newOpts = *opts
	}
	newOpts.TaskQueue = req.ResponseTaskQueue
	ctx = workflow.WithActivityOptions(ctx, newOpts)
	return workflow.ExecuteActivity(ctx, ExecResponseName, resp)
}

// SqliteChild executes a child workflow.
// If options not present, they are taken from the context.
func SqliteChild(ctx workflow.Context, opts *workflow.ChildWorkflowOptions, req *SqliteOptions) SqliteChildRun {
	if opts == nil {
		ctxOpts := workflow.GetChildWorkflowOptions(ctx)
		opts = &ctxOpts
	}
	ctx = workflow.WithChildOptions(ctx, *opts)
	return SqliteChildRun{workflow.ExecuteChildWorkflow(ctx, SqliteName, req)}
}

// SqliteChildRun is a future for the child workflow.
type SqliteChildRun struct{ Future workflow.ChildWorkflowFuture }

// WaitStart waits for the child workflow to start.
func (r SqliteChildRun) WaitStart(ctx workflow.Context) (*workflow.Execution, error) {
	var exec workflow.Execution
	if err := r.Future.GetChildWorkflowExecution().Get(ctx, &exec); err != nil {
		return nil, err
	}
	return &exec, nil
}

// SelectStart adds waiting for start to the selector. Callback can be nil.
func (r SqliteChildRun) SelectStart(sel workflow.Selector, fn func(SqliteChildRun)) workflow.Selector {
	return sel.AddFuture(r.Future.GetChildWorkflowExecution(), func(workflow.Future) {
		if fn != nil {
			fn(r)
		}
	})
}

// Get returns the completed workflow value, waiting if necessary.
func (r SqliteChildRun) Get(ctx workflow.Context) error {
	return r.Future.Get(ctx, nil)
}

// Select adds this completion to the selector. Callback can be nil.
func (r SqliteChildRun) Select(sel workflow.Selector, fn func(SqliteChildRun)) workflow.Selector {
	return sel.AddFuture(r.Future, func(workflow.Future) {
		if fn != nil {
			fn(r)
		}
	})
}

// Errors fail the workflow by default
func (r SqliteChildRun) Update(ctx workflow.Context, req *UpdateRequest) workflow.Future {
	return r.Future.SignalChildWorkflow(ctx, UpdateName, req)
}

func (r SqliteChildRun) Exec(ctx workflow.Context, req *ExecRequest) (ExecResponseExternal, error) {
	var resp ExecResponseExternal
	if req.Id == "" {
		return resp, fmt.Errorf("missing request ID")
	}
	if req.ResponseTaskQueue != "" {
		return resp, fmt.Errorf("cannot have task queue for child")
	}
	req.ResponseWorkflowId = workflow.GetInfo(ctx).WorkflowExecution.ID
	resp.Channel = workflow.GetSignalChannel(ctx, ExecResponseName+"-"+req.Id)
	resp.Future = r.Future.SignalChildWorkflow(ctx, ExecSignalName, req)
	return resp, nil
}

// Errors fail the workflow by default
func UpdateExternal(ctx workflow.Context, workflowID, runID string, req *UpdateRequest) workflow.Future {
	return workflow.SignalExternalWorkflow(ctx, workflowID, runID, UpdateName, req)
}

func ExecExternal(ctx workflow.Context, workflowID, runID string, req *ExecRequest) (ExecResponseExternal, error) {
	var resp ExecResponseExternal
	if req.Id == "" {
		return resp, fmt.Errorf("missing request ID")
	}
	if req.ResponseTaskQueue != "" {
		return resp, fmt.Errorf("cannot have task queue for child")
	}
	req.ResponseWorkflowId = workflow.GetInfo(ctx).WorkflowExecution.ID
	resp.Channel = workflow.GetSignalChannel(ctx, ExecResponseName+"-"+req.Id)
	resp.Future = workflow.SignalExternalWorkflow(ctx, workflowID, runID, ExecSignalName, req)
	return resp, nil
}

// ExecResponseExternal represents a call response.
type ExecResponseExternal struct {
	Future  workflow.Future
	Channel workflow.ReceiveChannel
}

// WaitSent blocks until the request is sent.
func (e ExecResponseExternal) WaitSent(ctx workflow.Context) error {
	return e.Future.Get(ctx, nil)
}

// SelectSent adds when a request is sent to the selector. Callback can be nil.
func (e ExecResponseExternal) SelectSent(sel workflow.Selector, fn func(ExecResponseExternal)) workflow.Selector {
	return sel.AddFuture(e.Future, func(workflow.Future) {
		if fn != nil {
			fn(e)
		}
	})
}

// Receive blocks until response is received.
func (e ExecResponseExternal) Receive(ctx workflow.Context) *ExecResponse {
	var resp ExecResponse
	e.Channel.Receive(ctx, &resp)
	return &resp
}

// ReceiveAsync returns response or nil if none.
func (e ExecResponseExternal) ReceiveAsync() *ExecResponse {
	var resp ExecResponse
	if !e.Channel.ReceiveAsync(&resp) {
		return nil
	}
	return &resp
}

// Select adds the callback to the selector to be invoked when response received. Callback can be nil
func (e ExecResponseExternal) Select(sel workflow.Selector, fn func(*ExecResponse)) workflow.Selector {
	return sel.AddReceive(e.Channel, func(workflow.ReceiveChannel, bool) {
		req := e.ReceiveAsync()
		if fn != nil {
			fn(req)
		}
	})
}
