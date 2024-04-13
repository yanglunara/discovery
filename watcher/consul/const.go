package consul

type DataCenter string

const (
	SingleDataCenter DataCenter = "SINGLE"
	MultiDataCenter  DataCenter = "MULTI"
)

func (d DataCenter) String() string {
	return string(d)
}
