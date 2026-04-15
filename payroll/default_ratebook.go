package payroll

// defaultRateBookData is the embedded T4127-JAN-2026 (122nd edition) tables.
var defaultRateBookData = RateBook{
	Edition: Edition,
	TaxYear: 2026,
	Federal: FederalSchedule{
		Bands: []FederalBand{
			{58523, 0.14, 0},
			{117045, 0.205, 3804},
			{181440, 0.26, 10241},
			{258482, 0.29, 15685},
			{1e15, 0.33, 26024},
		},
	},
	Provincial: map[Province]ProvincialBracketSet{
		AB: {Thresholds: []float64{0, 61200, 154259, 185111, 246813, 370220}, V: []float64{0.08, 0.10, 0.12, 0.13, 0.14, 0.15}, KP: []float64{0, 1224, 4309, 6160, 8628, 12331}},
		BC: {Thresholds: []float64{0, 50363, 100728, 115648, 140430, 190405, 265545}, V: []float64{0.0506, 0.0770, 0.1050, 0.1229, 0.1470, 0.1680, 0.2050}, KP: []float64{0, 1330, 4150, 6220, 9604, 13603, 23428}},
		MB: {Thresholds: []float64{0, 47000, 100000}, V: []float64{0.1080, 0.1275, 0.1740}, KP: []float64{0, 917, 5567}},
		NB: {Thresholds: []float64{0, 52333, 104666, 193861}, V: []float64{0.0940, 0.1400, 0.1600, 0.1950}, KP: []float64{0, 2407, 4501, 11286}},
		NL: {Thresholds: []float64{0, 44678, 89354, 159528, 223340, 285319, 570638, 1141275}, V: []float64{0.0870, 0.1450, 0.1580, 0.1780, 0.1980, 0.2080, 0.2130, 0.2180}, KP: []float64{0, 2591, 3753, 6943, 11410, 14263, 17117, 22823}},
		NS: {Thresholds: []float64{0, 30995, 61991, 97417, 157124}, V: []float64{0.0879, 0.1495, 0.1667, 0.1750, 0.2100}, KP: []float64{0, 1909, 2976, 3784, 9283}},
		NT: {Thresholds: []float64{0, 53003, 106009, 172346}, V: []float64{0.0590, 0.0860, 0.1220, 0.1405}, KP: []float64{0, 1431, 5247, 8436}},
		NU: {Thresholds: []float64{0, 55801, 111602, 181439}, V: []float64{0.0400, 0.0700, 0.0900, 0.1150}, KP: []float64{0, 1674, 3906, 8442}},
		ON: {Thresholds: []float64{0, 53891, 107785, 150000, 220000}, V: []float64{0.0505, 0.0915, 0.1116, 0.1216, 0.1316}, KP: []float64{0, 2210, 4376, 5876, 8076}},
		PE: {Thresholds: []float64{0, 33928, 65820, 106890, 142250}, V: []float64{0.0950, 0.1347, 0.1660, 0.1762, 0.1900}, KP: []float64{0, 1347, 3407, 4497, 6460}},
		SK: {Thresholds: []float64{0, 54532, 155805}, V: []float64{0.1050, 0.1250, 0.1450}, KP: []float64{0, 1091, 4207}},
		YT: {Thresholds: []float64{0, 58523, 117045, 181440, 500000}, V: []float64{0.0640, 0.0900, 0.1090, 0.1280, 0.1500}, KP: []float64{0, 1522, 3745, 7193, 18193}},
	},
	CPP: CPPParams{
		YMPE:                 74600.00,
		YAMPE:                85000.00,
		YMCE:                 71100.00,
		TotalRate:            0.0595,
		BaseRate:             0.0495,
		FirstAddRate:         0.0100,
		SecondAddRate:        0.0400,
		MaxTotalEmployee:     4230.45,
		MaxBaseEmployee:      3519.45,
		MaxFirstAdditional:   711.00,
		MaxSecondAdditional:  416.00,
		BasicExemptionAnnual: 3500.00,
	},
	EI: EIParams{
		MaxInsurableEarnings: 68900.00,
		RateCanadaExceptQC:   0.0163,
		RateQC:               0.0130,
		MaxEmployeeCanada:    1123.07,
		MaxEmployeeQC:        895.70,
	},
	FederalLowestRate:   0.14,
	CEA:                 1501.00,
	AlbertaK5PThreshold: 4896.00,
	BPAFederal: BPAFederalParams{
		Tier1MaxNI:           181440,
		Tier2MaxNI:           258482,
		BPAFull:              16452,
		BPAReduced:           14829,
		ReductionNumerator:   1623,
		ReductionDenominator: 77042,
	},
	MBParams: ManitobaBPAMBParams{
		Tier1MaxNI: 200000,
		Tier2MaxNI: 400000,
		FullAmount: 15780,
		Tier2Span:  200000,
	},
	TCPDefaults: TCPDefaults{
		AB: 22769, BC: 13216, NB: 13664, NL: 11188, NS: 11932,
		NT: 18198, NU: 19659, ON: 12989, PE: 15000, SK: 20381,
	},
}

func init() {
	if err := defaultRateBookData.Validate(); err != nil {
		panic(err)
	}
}
