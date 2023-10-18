package icalutil 

import (
	icaltypes "github.com/mattismoel/icalendar/types"
	"github.com/mattismoel/lectigo/types"
)

func ModuleToICalEvent(module *types.Module) *icaltypes.ICalEvent {
	return &icaltypes.ICalEvent{
		UID: module.Id,
		StartDate: module.StartDate,
		EndDate: module.EndDate,
		Summary: module.Title,
		Location: module.Room,
		Description: module.Homework,
	}
}

