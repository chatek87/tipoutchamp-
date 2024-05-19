package main

type Calculator struct {
	// input fields
	BarTeamIn BarTeamIn
	ServersIn []ServerIn
	EventsIn  []EventIn
	SupportIn []SupportIn

	// output fields
	BarTeamOut BarTeamOut
	ServersOut []ServerOut
	EventsOut  []EventOut
	SupportOut []SupportOut

	// configuration fields
	BarCount                int
	SupportCount            int
	TotalBarHours           float64
	TotalSupportHours       float64
	BarTipoutPercentage     float64
	SupportTipoutPercentage float64

	// tip pool-related fields
	BarPool     float64
	SupportPool float64

	ToSupportFromBarAmount     float64
	ToSupportFromServersAmount float64
	ToSupportFromEventsAmount  float64
}

func (c *Calculator) RunCalculationsPopulateOutputFields() {
	c.copyInputIntoOutput()
	c.setConfigurationFields()
	c.tallyTipPools()
	c.distributeTipoutsGetFinalPayouts()
}

// helpers to RunCalculationsPopulateOutputFields()
func (c *Calculator) copyInputIntoOutput() {
	// initialize output slices w/ same length as input slices
	c.BarTeamOut.Bartenders = make([]BartenderOut, len(c.BarTeamIn.Bartenders))
	c.ServersOut = make([]ServerOut, len(c.ServersIn))
	c.EventsOut = make([]EventOut, len(c.EventsIn))
	c.SupportOut = make([]SupportOut, len(c.SupportIn))
	// bar
	c.BarTeamOut.OwedToPreTipout = c.BarTeamIn.OwedTo
	c.BarTeamOut.Sales = c.BarTeamIn.Sales

	if c.BarTeamIn.Bartenders != nil {
		for i, bartender := range c.BarTeamIn.Bartenders {
			c.BarTeamOut.Bartenders[i].Name = bartender.Name
			c.BarTeamOut.Bartenders[i].Hours = bartender.Hours
		}
	}
	// servers
	if c.ServersIn != nil {
		for i, server := range c.ServersIn {
			c.ServersOut[i].Name = server.Name
			c.ServersOut[i].Sales = server.Sales
			c.ServersOut[i].OwedToPreTipout = server.OwedTo
		}
	}
	// events
	if c.EventsIn != nil {
		for i, event := range c.EventsIn {
			c.EventsOut[i].Name = event.Name
			c.EventsOut[i].OwedToPreTipout = event.OwedTo
			c.EventsOut[i].Sales = event.Sales
			c.EventsOut[i].SplitBy = event.SplitBy
		}
	}
	// support
	if c.SupportIn != nil {
		for i, support := range c.SupportIn {
			c.SupportOut[i].Name = support.Name
			c.SupportOut[i].Hours = support.Hours
		}
	}
}
func (c *Calculator) setConfigurationFields() {
	// counts
	c.BarCount = len(c.BarTeamIn.Bartenders)
	c.SupportCount = len(c.SupportIn)
	// tipout %'s
	c.setBarTipoutPercentage()
	c.setSupportTipoutPercentage()
	// hours
	c.setTotalBarHours()
	c.setTotalSupportHours()
}
func (c *Calculator) setTotalBarHours() {
	totalHours := 0.0
	for _, bartender := range c.BarTeamIn.Bartenders {
		totalHours += bartender.Hours
	}
	c.TotalBarHours = totalHours
}
func (c *Calculator) setTotalSupportHours() {
	totalHours := 0.0
	for _, support := range c.SupportIn {
		totalHours += support.Hours
	}
	c.TotalSupportHours = totalHours
}
func (c *Calculator) setBarTipoutPercentage() {
	count := len(c.SupportIn)
	if count >= 3 {
		c.BarTipoutPercentage = 0.015
	} else {
		c.BarTipoutPercentage = 0.02
	}
}
func (c *Calculator) setSupportTipoutPercentage() {
	count := len(c.SupportIn)
	if count == 0 {
		c.SupportTipoutPercentage = 0.00
	}
	if count <= 3 {
		c.SupportTipoutPercentage = float64(count) * 0.01
	} else {
		c.SupportTipoutPercentage = 0.03
	}
}
func (c *Calculator) tallyTipPools() {
	// bar pool
	// from servers
	for i := range c.ServersOut {
		server := &c.ServersOut[i]
		server.TipoutToBar = server.Sales * c.BarTipoutPercentage
		c.BarPool += server.TipoutToBar
		c.BarTeamOut.TipoutFromServers += server.TipoutToBar
	}
	// from events
	for i := range c.EventsOut {
		event := &c.EventsOut[i]
		event.TipoutToBar = event.Sales * c.BarTipoutPercentage
		c.BarPool += event.TipoutToBar
		c.BarTeamOut.TipoutFromEvents += event.TipoutToBar
	}

	c.BarTeamOut.TotalTipoutReceived = c.BarTeamOut.TipoutFromServers + c.BarTeamOut.TipoutFromEvents

	// support pool
	if c.SupportOut != nil {
		// from bar
		// calculate tipout to support and record in field
		c.BarTeamOut.TipoutToSupport = c.BarTeamOut.Sales * c.SupportTipoutPercentage
		// add it to the support pool running tally
		c.SupportPool += c.BarTeamOut.TipoutToSupport
		// vv kind of a redundant field now, but could be useful should tipout rules change
		c.BarTeamOut.TotalAmountTippedOut = c.BarTeamOut.TipoutToSupport
		c.ToSupportFromBarAmount += c.BarTeamOut.TipoutToSupport

		// from servers
		for i := range c.ServersOut {
			server := &c.ServersOut[i]
			server.TipoutToSupport = server.Sales * c.SupportTipoutPercentage
			c.SupportPool += server.TipoutToSupport
			c.ToSupportFromServersAmount += server.TipoutToSupport
		}
		// from events
		for i := range c.EventsOut {
			event := &c.EventsOut[i]
			event.TipoutToSupport = event.Sales * c.SupportTipoutPercentage
			c.SupportPool += event.TipoutToSupport
			c.ToSupportFromEventsAmount += event.TipoutToSupport
		}
	}
}
func (c *Calculator) distributeTipoutsGetFinalPayouts() {
	// bar team
	c.BarTeamOut.FinalPayout = c.BarTeamOut.OwedToPreTipout - c.BarTeamOut.TotalAmountTippedOut + c.BarPool
	// bartenders
	if c.BarTeamOut.Bartenders != nil {
		// for _, bartender := range c.BarTeamOut.Bartenders {
		for i := range c.BarTeamOut.Bartenders {
			bartender := &c.BarTeamOut.Bartenders[i]
			bartender.PercentageOfBarTipPool = bartender.Hours / c.TotalBarHours
			bartender.OwedToPreTipout = c.BarTeamOut.OwedToPreTipout * bartender.PercentageOfBarTipPool
			bartender.TipoutToSupport = c.BarTeamOut.TipoutToSupport * bartender.PercentageOfBarTipPool
			bartender.TotalAmountTippedOut = bartender.TipoutToSupport
			bartender.TipoutFromServers = c.BarTeamOut.TipoutFromServers * bartender.PercentageOfBarTipPool
			bartender.TipoutFromEvents = c.BarTeamOut.TipoutFromEvents * bartender.PercentageOfBarTipPool
			bartender.TotalTipoutReceived = bartender.TipoutFromServers + bartender.TipoutFromEvents
			bartender.FinalPayout = bartender.OwedToPreTipout - bartender.TotalAmountTippedOut + (c.BarPool * bartender.PercentageOfBarTipPool)
		}
	}
	// servers
	for i := range c.ServersOut {
		server := &c.ServersOut[i]
		server.TotalAmountTippedOut = server.TipoutToBar + server.TipoutToSupport
		server.FinalPayout = server.OwedToPreTipout - server.TotalAmountTippedOut
	}
	// events
	for i := range c.EventsOut {
		event := &c.EventsOut[i]
		event.TotalAmountTippedOut = event.TipoutToBar + event.TipoutToSupport
		event.FinalPayout = event.OwedToPreTipout - event.TotalAmountTippedOut
		event.FinalPayoutPerWorker = event.FinalPayout / float64(event.SplitBy)
	}
	// support
	for i := range c.SupportOut {
		support := &c.SupportOut[i]
		support.PercentageOfSupportTipPool = support.Hours / c.TotalSupportHours
		support.TipoutFromBar = c.ToSupportFromBarAmount * support.PercentageOfSupportTipPool
		support.TipoutFromServers = c.ToSupportFromServersAmount * support.PercentageOfSupportTipPool
		support.TipoutFromEvents = c.ToSupportFromEventsAmount * support.PercentageOfSupportTipPool
		support.FinalPayout = c.SupportPool * support.PercentageOfSupportTipPool
	}
}

// report related
// func (c *Calculator) GenerateReport() {

// }

// func (c *Calculator) SaveJSONToFile() {

// }
