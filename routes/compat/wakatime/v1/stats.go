package v1

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	conf "github.com/muety/wakapi/config"
	"github.com/muety/wakapi/middlewares"
	"github.com/muety/wakapi/models"
	v1 "github.com/muety/wakapi/models/compat/wakatime/v1"
	"github.com/muety/wakapi/services"
	"github.com/muety/wakapi/utils"
)

type StatsHandler struct {
	config      *conf.Config
	userSrvc    services.IUserService
	summarySrvc services.ISummaryService
}

func NewStatsHandler(userService services.IUserService, summaryService services.ISummaryService) *StatsHandler {
	return &StatsHandler{
		userSrvc:    userService,
		summarySrvc: summaryService,
		config:      conf.Get(),
	}
}

func (h *StatsHandler) RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("").Subrouter()
	r.Use(
		middlewares.NewAuthenticateMiddleware(h.userSrvc).WithOptionalFor([]string{"/"}).Handler,
	)
	r.Path("/v1/users/{user}/stats/{range}").Methods(http.MethodGet).HandlerFunc(h.Get)
	r.Path("/compat/wakatime/v1/users/{user}/stats/{range}").Methods(http.MethodGet).HandlerFunc(h.Get)

	// Also works without range, see https://github.com/anuraghazra/github-readme-stats/issues/865#issuecomment-776186592
	r.Path("/v1/users/{user}/stats").Methods(http.MethodGet).HandlerFunc(h.Get)
	r.Path("/compat/wakatime/v1/users/{user}/stats").Methods(http.MethodGet).HandlerFunc(h.Get)
}

// TODO: support filtering (requires https://github.com/muety/wakapi/issues/108)

// @Summary Retrieve statistics for a given user
// @Description Mimics https://wakatime.com/developers#stats
// @ID get-wakatimes-tats
// @Tags wakatime
// @Produce json
// @Param user path string true "User ID to fetch data for (or 'current')"
// @Param range path string false "Range interval identifier" Enums(today, yesterday, week, month, year, 7_days, last_7_days, 30_days, last_30_days, 12_months, last_12_months, any)
// @Param project query string false "Project to filter by"
// @Param language query string false "Language to filter by"
// @Param editor query string false "Editor to filter by"
// @Param operating_system query string false "OS to filter by"
// @Param machine query string false "Machine to filter by"
// @Param label query string false "Project label to filter by"
// @Security ApiKeyAuth
// @Success 200 {object} v1.StatsViewModel
// @Router /compat/wakatime/v1/users/{user}/stats/{range} [get]
func (h *StatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	var vars = mux.Vars(r)
	var authorizedUser, requestedUser *models.User

	authorizedUser = middlewares.GetPrincipal(r)
	if authorizedUser != nil && vars["user"] == "current" {
		vars["user"] = authorizedUser.ID
	}

	requestedUser, err := h.userSrvc.GetUserById(vars["user"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("user not found"))
		return
	}

	rangeParam := vars["range"]
	if rangeParam == "" {
		rangeParam = (*models.IntervalPast7Days)[0]
	}

	err, rangeFrom, rangeTo := utils.ResolveIntervalRawTZ(rangeParam, requestedUser.TZ())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid range"))
		return
	}

	minStart := rangeTo.Add(-24 * time.Hour * time.Duration(requestedUser.ShareDataMaxDays))
	if (authorizedUser == nil || requestedUser.ID != authorizedUser.ID) &&
		rangeFrom.Before(minStart) && requestedUser.ShareDataMaxDays >= 0 {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("requested time range too broad"))
		return
	}

	summary, err, status := h.loadUserSummary(requestedUser, rangeFrom, rangeTo, utils.ParseSummaryFilters(r))
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(err.Error()))
		return
	}

	stats := v1.NewStatsFrom(summary, &models.Filters{})

	// post filter stats according to user's given sharing permissions
	if !requestedUser.ShareEditors {
		stats.Data.Editors = nil
	}
	if !requestedUser.ShareLanguages {
		stats.Data.Languages = nil
	}
	if !requestedUser.ShareProjects {
		stats.Data.Projects = nil
	}
	if !requestedUser.ShareOSs {
		stats.Data.OperatingSystems = nil
	}
	if !requestedUser.ShareMachines {
		stats.Data.Machines = nil
	}

	utils.RespondJSON(w, r, http.StatusOK, stats)
}

func (h *StatsHandler) loadUserSummary(user *models.User, start, end time.Time, filters *models.Filters) (*models.Summary, error, int) {
	overallParams := &models.SummaryParams{
		From:      start,
		To:        end,
		User:      user,
		Recompute: false,
	}

	summary, err := h.summarySrvc.Aliased(overallParams.From, overallParams.To, user, h.summarySrvc.Retrieve, filters, false)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	return summary, nil, http.StatusOK
}
