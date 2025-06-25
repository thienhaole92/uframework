package reconws_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/thienhaole92/uframework/reconws"
)

//nolint:tagliatelle
type ExecutionReport struct {
	EventType               string `json:"e"`
	EventTime               int64  `json:"E"`
	TransactionTime         int64  `json:"T"`
	Symbol                  string `json:"s"`
	ClientOrderID           string `json:"c"`
	Side                    string `json:"S"`
	OrderType               string `json:"o"`
	TimeInForce             string `json:"f"`
	OrderQty                string `json:"q"`
	OrderPrice              string `json:"p"`
	OrderStopPrice          string `json:"P"`
	OrigClientOrderID       string `json:"C"`
	ExecType                string `json:"x"`
	OrderStatus             string `json:"X"`
	OrderID                 int64  `json:"i"`
	LastExecQty             string `json:"l"`
	FilledQty               string `json:"z"`
	LastExecPrice           string `json:"L"`
	Fee                     string `json:"n"`
	FeeCcy                  string `json:"N"`
	TradeID                 int64  `json:"t"`
	IsWorking               bool   `json:"w"`
	IsMaker                 bool   `json:"m"`
	Ignore                  bool   `json:"M"`
	CreateTime              int64  `json:"O"`
	CummulativeQuoteQty     string `json:"Z"`
	LastExecQuoteQty        string `json:"Y"`
	QuoteOrderQuantity      string `json:"Q"`
	WorkingTime             int64  `json:"W"`
	SelfTradePreventionMode string `json:"V"`
}

func (e ExecutionReport) FeeDecimal() decimal.Decimal {
	d, _ := decimal.NewFromString(e.Fee)

	return d
}

func (e ExecutionReport) LastExecPriceDecimal() decimal.Decimal {
	p, _ := decimal.NewFromString(e.LastExecPrice)

	return p
}

func (e ExecutionReport) LastExecQtyDecimal() decimal.Decimal {
	q, _ := decimal.NewFromString(e.LastExecQty)

	return q
}

func (e ExecutionReport) Amount() decimal.Decimal {
	p, _ := decimal.NewFromString(e.LastExecPrice)
	q, _ := decimal.NewFromString(e.LastExecQty)

	return p.Mul(q)
}

//nolint:lll
func Test_Unpack_Payload(t *testing.T) {
	t.Parallel()

	p := []byte(`{"e":"executionReport","E":1729844837277,"s":"FDUSDUSDT","c":"220dfe7a31_q_csdlcpb802uc739lbddg","S":"BUY","o":"LIMIT","f":"GTC","q":"893.00000000","p":"0.99960000","P":"0.00000000","F":"0.00000000","g":-1,"C":"","x":"TRADE","X":"FILLED","r":"NONE","i":309729695,"l":"893.00000000","z":"893.00000000","L":"0.99960000","n":"0.00000000","N":"BNB","T":1729844837277,"t":218621754,"I":838105751,"w":false,"m":false,"M":true,"O":1729844837277,"Z":"892.64280000","Y":"892.64280000","Q":"0.00000000","W":1729844837277,"V":"EXPIRE_TAKER"}`)

	pl := reconws.Payload(p)

	var evt ExecutionReport

	err := pl.Unpack(&evt)
	if err != nil {
		t.Errorf("want %s, got %s", "nil", err.Error())
	}

	if evt.IsMaker != false {
		t.Errorf("want %t, got %t", false, evt.IsMaker)
	}

	if evt.LastExecPrice != "0.99960000" {
		t.Errorf("want %s, got %s", "0.99960000", evt.LastExecPrice)
	}

	if evt.LastExecQty != "893.00000000" {
		t.Errorf("want %s, got %s", "893.00000000", evt.LastExecQty)
	}

	amount := decimal.NewFromFloat(892.6428)

	if evt.Amount().Cmp(amount) != 0 {
		t.Errorf("want %s, got %s", evt.Amount().String(), amount.String())
	}
}
