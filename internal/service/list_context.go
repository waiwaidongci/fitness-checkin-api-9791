package service

import (
	"context"

	"fitness-checkin-api/internal/repository"
)

func newListContextFilter(ctx context.Context, params ListParams, page, pageSize int) repository.ListFilter {
	if ctx == nil {
		ctx = context.Background()
	}

	return repository.ListFilter{
		SportType: normalizeSportType(params.SportType),
		StartDate: params.StartDate,
		EndDate:   params.EndDate,
		Page:      page,
		PageSize:  pageSize,
	}
}
