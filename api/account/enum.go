package account

// Type identifies an account type
type Type string

const (
	// TypeChecking identifies a checking account
	TypeChecking Type = "checking"
	// TypeSavings identifies a savings account
	TypeSavings Type = "savings"
	// TypeCash identifies a cash account
	TypeCash Type = "cash"
	// TypeCreditCard identifies a credit card account
	TypeCreditCard Type = "creditCard"
	// TypeLineOfCredit identifies a line of credit account
	TypeLineOfCredit Type = "lineOfCredit"
	// TypeOtherAsset identifies an other asset account
	TypeOtherAsset Type = "otherAsset"
	// TypeOtherLiability identifies an other liability account
	TypeOtherLiability Type = "otherLiability"
	// TypeMortgage identifies a mortgage account
	TypeMortgage Type = "mortgage"
	// TypeAutoLoan identifies an auto loan account
	TypeAutoLoan Type = "autoLoan"
	// TypeStudentLoan identifies a student loan account
	TypeStudentLoan Type = "studentLoan"
	// TypePersonalLoan identifies a personal loan account
	TypePersonalLoan Type = "personalLoan"
	// TypeMedicalDebt identifies a medical debt account
	TypeMedicalDebt Type = "medicalDebt"
	// TypeOtherDebt identifies an other debt account
	TypeOtherDebt Type = "otherDebt"
	// TypePayPal DEPRECATED identifies a PayPal account
	TypePayPal Type = "payPal"
	// TypeMerchant DEPRECATED identifies a merchant account
	TypeMerchant Type = "merchantAccount"
	// TypeInvestment DEPRECATED identifies an investment account
	TypeInvestment Type = "investmentAccount"
)
