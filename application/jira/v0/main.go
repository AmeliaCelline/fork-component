//go:generate compogen readme ./config ./README.mdx
package jira

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
)

const (
	apiBaseURL       = "https://api.atlassian.com"
	taskListBoards   = "TASK_LIST_BOARDS"
	taskListIssues   = "TASK_LIST_ISSUES"
	taskListSprints  = "TASK_LIST_SPRINTS"
	taskGetIssue     = "TASK_GET_ISSUE"
	taskGetSprint    = "TASK_GET_SPRINT"
	taskCreateIssue  = "TASK_CREATE_ISSUE"
	taskUpdateIssue  = "TASK_UPDATE_ISSUE"
	taskCreateSprint = "TASK_CREATE_SPRINT"
	taskUpdateSprint = "TASK_UPDATE_SPRINT"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/setup.json
	setupJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte

	once sync.Once
	comp *component
)

type component struct {
	base.Component
}

type execution struct {
	base.ComponentExecution
	execute func(context.Context, *structpb.Struct) (*structpb.Struct, error)
	client  Client
}

// Init returns an implementation of IConnector that interacts with Slack.
func Init(bc base.Component) *component {
	once.Do(func() {
		comp = &component{Component: bc}
		err := comp.LoadDefinition(definitionJSON, setupJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})

	return comp
}

func (c *component) CreateExecution(x base.ComponentExecution) (base.IExecution, error) {
	ctx := context.Background()
	jiraClient, err := newClient(ctx, x.Setup, c.Logger)
	if err != nil {
		return nil, err
	}
	e := &execution{
		ComponentExecution: x,
		client:             *jiraClient,
	}
	// docs: https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/#about
	switch x.Task {
	case taskListBoards:
		e.execute = e.client.listBoardsTask
	case taskListIssues:
		e.execute = e.client.listIssuesTask
	case taskListSprints:
		e.execute = e.client.listSprintsTask
	case taskGetIssue:
		e.execute = e.client.getIssueTask
	case taskGetSprint:
		e.execute = e.client.getSprintTask
	case taskCreateIssue:
		e.execute = e.client.createIssueTask
	case taskUpdateIssue:
		e.execute = e.client.updateIssueTask
	case taskCreateSprint:
		e.execute = e.client.createSprintTask
	case taskUpdateSprint:
		e.execute = e.client.updateSprintTask
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", x.Task),
			fmt.Sprintf("%s task is not supported.", x.Task),
		)
	}

	return e, nil
}

func (e *execution) Execute(ctx context.Context, in base.InputReader, out base.OutputWriter) error {
	inputs, err := in.Read(ctx)
	if err != nil {
		return err
	}
	outputs := make([]*structpb.Struct, len(inputs))

	for i, input := range inputs {
		input, err := e.ComponentExecution.FillInDefaultValues(input)
		if err != nil {
			return err
		}
		output, err := e.execute(ctx, input)
		if err != nil {
			return err
		}

		outputs[i] = output
	}

	return out.Write(ctx, outputs)
}
