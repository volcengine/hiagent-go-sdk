package version

const (
	DefaultRegion = "cn-north-1"

	ServerService  = "hibot-server"
	GatewayService = "hibot-gateway"
	// AIGWService is the TOP DestService name for model/provider Actions.
	// Some deployments register these Actions under this exact service name;
	// "aigw-server" may return InvalidAction in those environments.
	AIGWService = "aigw"
	UPService   = "up"

	V1 = "v1"

	// Server / Chat / Model / UP are TOP-registered API Versions; the real
	// gateway PreCheck enforces the YYYY-MM-DD format and rejects "v1".
	//
	// Models go through the AIGW service (DestService=aigw), whose API
	// Version is pinned at 2023-08-01 for compatibility.
	Server = "2026-04-23"
	Chat   = "2026-05-11"
	Model  = "2023-08-01"
	UP     = "2022-01-01"
)
