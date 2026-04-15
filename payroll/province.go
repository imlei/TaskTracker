package payroll

// Province of employment for payroll deductions (excluding Quebec provincial tax in T4127 T2).
type Province int

const (
	AB Province = iota
	BC
	MB
	NB
	NL
	NS
	NT
	NU
	ON
	PE
	QC
	SK
	YT
	OutsideCanada
)

func (p Province) String() string {
	switch p {
	case AB:
		return "AB"
	case BC:
		return "BC"
	case MB:
		return "MB"
	case NB:
		return "NB"
	case NL:
		return "NL"
	case NS:
		return "NS"
	case NT:
		return "NT"
	case NU:
		return "NU"
	case ON:
		return "ON"
	case PE:
		return "PE"
	case QC:
		return "QC"
	case SK:
		return "SK"
	case YT:
		return "YT"
	case OutsideCanada:
		return "OUTSIDE_CA"
	default:
		return "UNKNOWN"
	}
}
