package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	alcsoap "webctrl-soap-go/pkg/webctrl_soap_go"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type AppConfig struct {
	AWSConfig aws.Config
}

var appConfig AppConfig

func handler(ctx context.Context, req LambdaRequest) (LambdaResponse, error) {
	if err := validateRequest(req); err != nil {
		return LambdaResponse{}, err
	}

	requestedBy := strings.TrimSpace(req.RequestedBy)
	if requestedBy == "" {
		requestedBy = "automation"
	}

	server := strings.TrimRight(strings.TrimSpace(req.Server), "/")

	password, err := loadPasswordFromSSM(ctx, appConfig.AWSConfig, req.PasswordSsmPath)
	if err != nil {
		return LambdaResponse{}, err
	}

	httpClient, err := alcsoap.NewHTTPClientWithExtraCAPEM(root1, root2, root3)
	if err != nil {
		return LambdaResponse{}, fmt.Errorf("create WebCTRL HTTP client: %w", err)
	}

	alc := alcsoap.NewSoapServiceWithHTTPClient(
		server,
		strings.TrimSpace(req.Username),
		password,
		httpClient,
	)

	resp := LambdaResponse{
		JobName:         req.JobName,
		RequestedBy:     requestedBy,
		WritesRequested: len(req.Writes),
		Results:         make([]WriteResult, 0, len(req.Writes)),
	}

	for _, write := range req.Writes {
		result := WriteResult{
			Path:   write.Path,
			Value:  write.Value,
			Status: "success",
		}

		value, err := gqlValueToString(write.Value)
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			resp.WritesFailed++
			resp.Results = append(resp.Results, result)
			continue
		}

		err = alc.Eval.SetValue(write.Path, value, requestedBy)
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			resp.WritesFailed++
			resp.Results = append(resp.Results, result)
			continue
		}

		resp.WritesSucceeded++
		resp.Results = append(resp.Results, result)
	}

	return resp, nil
}

func main() {
	cfg, err := loadAppConfig(context.Background())
	if err != nil {
		panic(fmt.Errorf("load app config: %w", err))
	}

	appConfig = cfg

	lambda.Start(handler)
}

func loadAppConfig(ctx context.Context) (AppConfig, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return AppConfig{}, fmt.Errorf("load AWS config: %w", err)
	}

	return AppConfig{
		AWSConfig: awsCfg,
	}, nil
}

func validateRequest(req LambdaRequest) error {
	if strings.TrimSpace(req.JobName) == "" {
		return errors.New("missing jobName")
	}

	if strings.TrimSpace(req.Server) == "" {
		return errors.New("missing server")
	}

	if strings.TrimSpace(req.Username) == "" {
		return errors.New("missing username")
	}

	if strings.TrimSpace(req.PasswordSsmPath) == "" {
		return errors.New("missing passwordSsmPath")
	}

	if len(req.Writes) == 0 {
		return errors.New("writes must contain at least one write")
	}

	for i, write := range req.Writes {
		if strings.TrimSpace(write.Path) == "" {
			return fmt.Errorf("writes[%d].path is required", i)
		}

		if len(write.Value) == 0 {
			return fmt.Errorf("writes[%d].value is required", i)
		}
	}

	return nil
}

func loadPasswordFromSSM(ctx context.Context, awsCfg aws.Config, path string) (string, error) {
	path = strings.TrimSpace(path)

	ssmClient := ssm.NewFromConfig(awsCfg)

	out, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(path),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("get SSM parameter %s: %w", path, err)
	}

	if out.Parameter == nil || out.Parameter.Value == nil || *out.Parameter.Value == "" {
		return "", fmt.Errorf("missing SSM parameter value %s", path)
	}

	return *out.Parameter.Value, nil
}

func gqlValueToString(raw json.RawMessage) (string, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, nil
	}

	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		if b {
			return "1", nil
		}
		return "0", nil
	}

	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return strconv.FormatFloat(f, 'f', -1, 64), nil
	}

	return "", fmt.Errorf("unsupported value type: %s", string(raw))
}
