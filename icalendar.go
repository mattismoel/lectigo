package lectigo

import "github.com/mattismoel/icalendar"

func ModuleToICalEvent(module *Module) *icalendar.ICalEvent {
	return &icalendar.ICalEvent{
		UID: module.Id,
		StartDate: module.StartDate,
		EndDate: module.EndDate,
		Summary: module.Title,
		Location: module.Room,
		Description: module.Homework,
	}
}

