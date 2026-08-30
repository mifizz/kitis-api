package watcher

import (
	"context"
	"kitis-api/internal/database"
	"kitis-api/internal/domain"
	"reflect"
)

func compareSchedule(ctx context.Context, new domain.Schedule) (bool, error) {
	old, err := database.GetScheduleNewest(ctx, new.SourceType, new.Source)
	if err != nil {
		return false, err
	}

	return reflect.DeepEqual(old.Days, new.Days), nil
}
