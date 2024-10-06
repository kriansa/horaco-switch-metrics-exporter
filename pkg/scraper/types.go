package scraper

type PortStats struct {
	Enabled          bool
	Connected        bool
	TxGood           int
	TxBad            int
	RxGood           int
	RxBad            int
	Speed            LinkSpeed
	TransmissionMode LinkTransmissionMode
}

type LinkSpeed string

const (
	LinkSpeed10Mbps  LinkSpeed = "10 Mbps"
	LinkSpeed100Mbps LinkSpeed = "100 Mbps"
	LinkSpeed1Gbps   LinkSpeed = "1 Gbps"
	LinkSpeed2_5Gbps LinkSpeed = "2.5 Gbps"
	LinkSpeed10Gbps  LinkSpeed = "10 Gbps"
)

type LinkTransmissionMode string

const (
	LinkTransmissionModeHalfDuplex LinkTransmissionMode = "Half Duplex"
	LinkTransmissionModeFullDuplex LinkTransmissionMode = "Full Duplex"
)
