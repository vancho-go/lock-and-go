package userdata

import (
	"context"
	"errors"
	"github.com/vancho-go/lock-and-go/internal/model"
	"github.com/vancho-go/lock-and-go/internal/repository/storage/psql"
	"github.com/vancho-go/lock-and-go/internal/service/jwt"
)

type DataService struct {
	upserter psql.UserDataUpserter
	deleter  psql.UserDataDeleter
	reader   psql.UserDataReader
}

func NewDataService(
	upserter psql.UserDataUpserter,
	reader psql.UserDataReader,
	deleter psql.UserDataDeleter) *DataService {
	return &DataService{
		upserter: upserter,
		reader:   reader,
		deleter:  deleter,
	}
}

func (s *DataService) SyncDataChanges(ctx context.Context, datas []model.UserData) error {
	userID, ok := jwt.GetUserIDFromContext(ctx)
	if !ok {
		return errors.New("user ID not found in context")
	}

	// Разделяем данные на те, что нужно обновить/добавить и на те, что нужно удалить
	var toUpsert, toDelete []model.UserData
	for _, data := range datas {
		data.UserID = userID
		if data.Status == "deleted" {
			toDelete = append(toDelete, data)
		} else {
			toUpsert = append(toUpsert, data)
		}
	}

	if len(toUpsert) > 0 {
		if err := s.upserter.Upsert(ctx, toUpsert); err != nil {
			return err
		}
	}

	if len(toDelete) > 0 {
		if err := s.deleter.Delete(ctx, toDelete); err != nil {
			return err
		}
	}

	return nil
}

func (s *DataService) GetData(ctx context.Context) ([]model.UserData, error) {
	userID, ok := jwt.GetUserIDFromContext(ctx)
	if !ok {
		return nil, errors.New("user ID not found in context")
	}
	return s.reader.Read(ctx, userID)
}
