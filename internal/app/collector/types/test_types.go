package types

const BlunaTokenInfo = `{"height":"3754668","result":{"name":"Bonded Luna","symbol":"BLUNA","decimals":6,"total_supply":"79178685320809"}}`

const (
	HubContract                 = "terra1mtwph2juhj0rvjz7dy92gvl6xvukaxu8rfv8ts"
	RewardContract              = "terra17yap3mhph35pcwvhza38c2lkj7gzywzy05h7l0"
	BlunaTokenInfoContract      = "terra1kc87mu460fwkqte29rquh4hc20m54fxwtsx7gp"
	UpdateGlobalIndexBotAddress = "dummy_updateglobalindexbot"
	RewardDispatcherContract    = "dummy_rewarddispatcher"
	ValidatorsRegistryContract  = "dummy_validatorsregistry"
	AirDropRegistryContract     = "dummy_airdropRegistry"
)

const (
	TestValAddress   = "terravalcons1ezj3lps8nqwytt42at2sgt7seq9hk708g0spyk"
	TestValAddress2  = "terravaloper1qxqrtvg3smlfdfhvwcdzh0huh4f50kfs6gdt4x"
	TestValPublicKey = "terravalconspub1zcjduepqw2hyr7u7y70z5kdewn00xuq0wwcvnn0s7x5pjqcdpn80qsyctcpqcjhz4c"
	TestConsAddress  = "terravalcons1rfaxjug6md5jrz3c0uctyt6pzd50xyxlc2tf5m"
)

const (
	TestMoniker        = "Test validator"
	TestCommissionRate = 0.08
)

var (
	MissingVotesGovernanceLookbackLimit       = 10
	MissingVotesGovernanceAlertLimit          = 10
	MissingVotesGovernanceMonitoredValidators = []string{"terra17kj7euq7cyedllkm5z57svvq2quxmvzq9gxfy2", "terra1xal43l3e62an3q0mky99chn9ga4j5yhnfkegx8", "terra1lph93xlfjek830h0659p0wuewu56gkjayvjucj", "terra1spvnfmgklu6ynaph007h7zszmu5sgu97k7877c", "terra1zd2dg4y3734lfywkke40dvecu8nkzd2fxp8f0v", "terra1gu9h26qeyf4etuzke3h7qq6x9rjpw4xqkprntq", "terra1xqt30x2updtqseetay5c3r7znjqfltuwc68h82", "terra1qqu376azltyc5wnsje5qgwru5mtj2yqdhj0cwl", "terra1cdzstr5qhe9wvm4ec0jmpjgprv2hpuqkx9klzn", "terra1ya8mnvt0c6rahzvgj4ezwkt49rsscn6l8k9yhv", "terra1ya8mnvt0c6rahzvgj4ezwkt49rsscn6lkeklul"}
)
