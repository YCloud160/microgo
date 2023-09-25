// Code generated by protoc-gen-go-microgo. DO NOT EDIT.
// versions:
// - protoc-gen-go-microgo v1.0.0
// - protoc             v3.12.2
// source: greet.proto

package model

import (
	context "context"
	fmt "fmt"
	microgo "github.com/YCloud160/microgo"
	proto "google.golang.org/protobuf/proto"
)

type IGreetObjServer interface {
	SayHello(ctx context.Context, input *SayHelloReq) (output *SayHelloResp, err error)
}

type NopGreetObjServerImpl struct{}

func (*NopGreetObjServerImpl) SayHello(ctx context.Context, input *SayHelloReq) (output *SayHelloResp, err error) {
	return nil, fmt.Errorf("method SayHello not implement")
}

// GreetObjCall is used to call the implement of the defined method.
func GreetObjCall(ctx context.Context, impl any, enc microgo.Encoder, method string, input []byte) (out []byte, err error) {
	obj := impl.(IGreetObjServer)
	_ = obj
	switch method {
	case "SayHello":
		var req SayHelloReq
		if err = enc.Unmarshal(input, &req); err != nil {
			return nil, err
		}
		resp, err := obj.SayHello(ctx, &req)
		if err != nil {
			return nil, err
		}
		out, err = enc.Marshal(resp)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("method %s not implement", method)
	}
	return out, nil
}

// greetObjClient implement
type greetObjClient struct {
	client *microgo.Client
}

func NewGreetObjClient(name string, options ...microgo.ClientOption) *greetObjClient {
	client := microgo.NewClient(name, options...)
	return &greetObjClient{client: client}
}

func (client *greetObjClient) SayHello(ctx context.Context, req *SayHelloReq) (*SayHelloResp, error) {
	input, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	out, err := client.client.Call(ctx, "", "proto", "SayHello", input)
	if err != nil {
		return nil, err
	}
	resp := SayHelloResp{}
	if err := proto.Unmarshal(out, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
