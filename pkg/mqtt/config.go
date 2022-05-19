package mqtt

// Config : structure to hold MQTT configuration
type Config struct {
	HOST     string `yaml:"host" envconfig:"MQTT_BROKER_HOST"`
	PORT     uint   `yaml:"port" envconfig:"MQTT_BROKER_PORT"`
	USER     string `yaml:"user" envconfig:"MQTT_BROKER_USER"`
	PASSWORD string `yaml:"password" envconfig:"MQTT_BROKER_PASSWORD"`
	TOPICS   topics `yaml:"topics" envconfig:"TOPICS"`
}

type topics struct {
	RegistrationTopic string `yaml:"registration_topic" envconfig:"REGISTRATION_TOPIC"`
	RulesTopic        string `yaml:"rules_topic" envconfig:"RULES_TOPIC"`
	ActiveRulesTopic  string `yaml:"active_rules_topic" envconfig:"ACTIVE_RULES_TOPIC"`
	ActiveUsersTopic  string `yaml:"active_users_topic" envconfig:"ACTIVE_USERS_TOPIC"`
}

func (x *Config) Init() {
	x.HOST = "localhost"
	x.PORT = 1883
	x.USER = ""
	x.PASSWORD = ""
}
