package simulation

type Scenario struct {
	Name        string
	Description string
}

var scenarios = map[string]Scenario{
	"institution_impersonation": {
		Name:        "Institution Impersonation",
		Description: "A scenario where the user receives an call impersonating a trusted institution.",
	},
	"loan_scam": {
		Name:        "Loan Scam",
		Description: "A scenario where the user is targeted with a loan scam call.",
	},
	"delivery_notification": {
		Name:        "Delivery Notification",
		Description: "A scenario where the user receives a fake delivery notification call.",
	},
	"friends_impersonation": {
		Name:        "Friends Impersonation",
		Description: "A scenario where the user receives a call impersonating a friend in need.",
	},
}

func GetScenario(scenarioKey string) (Scenario, bool) {
	scenario, exists := scenarios[scenarioKey]
	return scenario, exists
}
