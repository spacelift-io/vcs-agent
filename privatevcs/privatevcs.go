package privatevcs

// MetadataFieldVcsAgentPoolULID is the gRPC metadata header where the agent sends its pool ulid.
const MetadataFieldVcsAgentPoolULID = "vcs-agent-pool-ulid"

// MetadataFieldVCSAgentKey is the gRPC metadata header where the agent sends the pool key.
const MetadataFieldVCSAgentKey = "vcs-agent-key"

// MetadataFieldVCSAgentMetadata is the gRPC metadata header where the agent sends its metadata.
const MetadataFieldVCSAgentMetadata = "vcs-agent-metadata"

// AgentPoolConfig is the configuration which an Agent uses to find a Gateway and authenticate as an Agent Pool.
type AgentPoolConfig struct {
	Host     string `json:"host"`
	Key      string `json:"key"`
	PoolULID string `json:"pool_ulid"`
}
