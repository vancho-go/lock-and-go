package handlers

import (
	"encoding/json"
	"github.com/vancho-go/lock-and-go/internal/model"
	userdata "github.com/vancho-go/lock-and-go/internal/service/user-data"
	"github.com/vancho-go/lock-and-go/pkg/logger"
	"net/http"
)

// UserDataController контроллер для обработки хранимых пользовательских данных.
type UserDataController struct {
	dataService *userdata.DataService
	log         *logger.Logger
}

// NewUserDataController конструктор UserDataController.
func NewUserDataController(service *userdata.DataService, log *logger.Logger) *UserDataController {
	return &UserDataController{
		dataService: service,
		log:         log}
}

// SyncDataChanges обработчик синхронизации пользовательских данных.
func (c *UserDataController) SyncDataChanges(w http.ResponseWriter, r *http.Request) {
	var datum []model.UserData
	if err := decodeJSONRequestBody(w, r, &datum); err != nil {
		return
	}

	if err := c.dataService.SyncDataChanges(r.Context(), datum); err != nil {
		c.log.Errorf("Failed to sync data changes: %v", err)
		http.Error(w, "Failed to sync data changes", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetData обработчик запроса на получение хранимых пользовательских данных.
func (c *UserDataController) GetData(w http.ResponseWriter, r *http.Request) {
	datum, err := c.dataService.GetData(r.Context())
	if err != nil {
		c.log.Errorf("Failed to get data: %v", err)
		http.Error(w, "Failed to get data", http.StatusInternalServerError)
		return
	}

	if len(datum) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(datum); err != nil {
		c.log.Errorf("Failed to encode data: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}
