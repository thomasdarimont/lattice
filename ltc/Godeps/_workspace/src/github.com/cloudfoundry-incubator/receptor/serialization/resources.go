package serialization

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	oldmodels "github.com/cloudfoundry-incubator/runtime-schema/models"
)

func EnvironmentVariablesToModel(envVars []receptor.EnvironmentVariable) []oldmodels.EnvironmentVariable {
	if envVars == nil {
		return nil
	}
	out := make([]oldmodels.EnvironmentVariable, len(envVars))
	for i, val := range envVars {
		out[i].Name = val.Name
		out[i].Value = val.Value
	}
	return out
}

func EnvironmentVariablesFromModel(envVars []oldmodels.EnvironmentVariable) []receptor.EnvironmentVariable {
	if envVars == nil {
		return nil
	}
	out := make([]receptor.EnvironmentVariable, len(envVars))
	for i, val := range envVars {
		out[i].Name = val.Name
		out[i].Value = val.Value
	}
	return out
}

func PortMappingFromProto(ports []*models.PortMapping) []receptor.PortMapping {
	if ports == nil {
		return nil
	}
	out := make([]receptor.PortMapping, len(ports))
	for i, val := range ports {
		out[i].ContainerPort = uint16(val.GetContainerPort())
		out[i].HostPort = uint16(val.GetHostPort())
	}
	return out
}

func PortMappingFromModel(ports []oldmodels.PortMapping) []receptor.PortMapping {
	if ports == nil {
		return nil
	}
	out := make([]receptor.PortMapping, len(ports))
	for i, val := range ports {
		out[i].ContainerPort = val.ContainerPort
		out[i].HostPort = val.HostPort
	}
	return out
}

func PortMappingToModel(ports []receptor.PortMapping) []oldmodels.PortMapping {
	if ports == nil {
		return nil
	}
	out := make([]oldmodels.PortMapping, len(ports))
	for i, val := range ports {
		out[i].ContainerPort = val.ContainerPort
		out[i].HostPort = val.HostPort
	}
	return out
}

