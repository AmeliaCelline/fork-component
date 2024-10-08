package instill

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

func (e *execution) executeImageClassification(grpcClient modelPB.ModelPublicServiceClient, nsID string, modelID string, version string, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	if len(inputs) <= 0 {
		return nil, fmt.Errorf("invalid input: %v for model: %s/%s/%s", inputs, nsID, modelID, version)
	}

	if grpcClient == nil {
		return nil, fmt.Errorf("uninitialized client")
	}

	taskInputs := []*modelPB.TaskInput{}
	for _, input := range inputs {

		classificationInput := &modelPB.ClassificationInput{}
		classificationInput.Type = &modelPB.ClassificationInput_ImageBase64{
			ImageBase64: base.TrimBase64Mime(input.Fields["image-base64"].GetStringValue()),
		}

		taskInput := &modelPB.TaskInput_Classification{
			Classification: classificationInput,
		}
		taskInputs = append(taskInputs, &modelPB.TaskInput{Input: taskInput})
	}

	taskOutputs, err := trigger(grpcClient, e.SystemVariables, nsID, modelID, version, taskInputs)
	if err != nil {
		return nil, err
	}
	if len(taskOutputs) <= 0 {
		return nil, fmt.Errorf("invalid output: %v for model: %s/%s/%s", taskOutputs, nsID, modelID, version)
	}
	outputs := []*structpb.Struct{}
	for idx := range inputs {
		imgClassificationOp := taskOutputs[idx].GetClassification()
		if imgClassificationOp == nil {
			return nil, fmt.Errorf("invalid output: %v for model: %s/%s/%s", imgClassificationOp, nsID, modelID, version)
		}
		outputJSON, err := protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}.Marshal(imgClassificationOp)
		if err != nil {
			return nil, err
		}
		output := &structpb.Struct{}
		err = protojson.Unmarshal(outputJSON, output)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}
