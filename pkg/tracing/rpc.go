package tracing

import (
	"context"
	"fmt"
	"github.com/AliyunContainerService/terway/rpc"
)

type tracingRPC struct {
	tracer *Tracer
}

func DefaultRPCServer() rpc.TerwayTracingServer {
	return RPCServer(&defaultTracer)
}

func RPCServer(tracer *Tracer) rpc.TerwayTracingServer {
	return &tracingRPC{tracer: tracer}
}

func (t *tracingRPC) GetResourceTypes(_ context.Context, _ *rpc.Placeholder) (*rpc.ResourcesTypesReply, error) {
	names := t.tracer.GetTypes()
	response := &rpc.ResourcesTypesReply{TypeNames: names}

	return response, nil
}

func (t *tracingRPC) GetResources(_ context.Context, request *rpc.ResourceTypeRequest) (*rpc.ResourcesNamesReply, error) {
	names := t.tracer.GetResourceNames(request.Name)
	response := &rpc.ResourcesNamesReply{ResourceNames: names}

	return response, nil
}

func (t *tracingRPC) GetResourceConfig(_ context.Context, request *rpc.ResourceTypeNameRequest) (*rpc.ResourceConfigReply, error) {
	config, err := t.tracer.GetConfig(request.Type, request.Name)
	if err != nil {
		return nil, err
	}

	var entry []*rpc.MapKeyValueEntry
	for _, v := range config {
		entry = append(entry, toRPCEntry(v))
	}

	response := &rpc.ResourceConfigReply{Config: entry}
	return response, nil
}

func (t *tracingRPC) GetResourceTrace(_ context.Context, request *rpc.ResourceTypeNameRequest) (*rpc.ResourceTraceReply, error) {
	trace, err := t.tracer.GetTrace(request.Type, request.Name)
	if err != nil {
		return nil, err
	}

	var entry []*rpc.MapKeyValueEntry
	for _, v := range trace {
		entry = append(entry, toRPCEntry(v))
	}

	response := &rpc.ResourceTraceReply{Trace: entry}
	return response, nil
}

func (t *tracingRPC) ResourceExecute(request *rpc.ResourceExecuteRequest, server rpc.TerwayTracing_ResourceExecuteServer) error {
	c, err := t.tracer.Execute(request.Type, request.Name, request.Command, request.Args)
	if err != nil {
		return err
	}

	for message := range c {
		err = server.Send(&rpc.ResourceExecuteReply{Message: message})
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *tracingRPC) GetResourceMapping(_ context.Context, _ *rpc.Placeholder) (*rpc.PodResourceMappingReply, error) {
	fmt.Println("get resource mapping from rpc")
	mapping, err := t.tracer.GetResourceMapping()
	if err != nil {
		return nil, err
	}

	var info []*rpc.PodResourceMapping
	for _, m := range mapping {
		info = append(info, toRPCMapping(m))
	}

	return &rpc.PodResourceMappingReply{
		Info: info,
	}, nil
}

func toRPCMapping(mapping PodResourceMapping) *rpc.PodResourceMapping {
	rMapping := rpc.PodResourceMapping{
		Type:                rpc.ResourceMappingType_MappingTypeNormal,
		PodName:             mapping.PodName,
		ResourceName:        mapping.Resource.ResID,
		FactoryResourceName: mapping.Resource.FactoryResource.ResID,
	}

	if !mapping.Valid {
		// FactoryResourceName == "" for case resource has been in the pool but not found via API
		if !mapping.Resource.Valid || rMapping.FactoryResourceName == "" {
			rMapping.Type = rpc.ResourceMappingType_MappingTypeError
		} else {
			rMapping.Type = rpc.ResourceMappingType_MappingTypeIdle
		}
	}

	return &rMapping
}

func toRPCEntry(entry MapKeyValueEntry) *rpc.MapKeyValueEntry {
	return &rpc.MapKeyValueEntry{
		Key:   entry.Key,
		Value: entry.Value,
	}
}
