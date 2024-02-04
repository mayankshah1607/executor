package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleTaskSubmission(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var tasks map[string]int32
	err := decoder.Decode(&tasks)
	if err != nil {
		s.log.Err(err).Msg("error decoding")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := s.service.AddTasks(tasks); err != nil {
		s.log.Err(err).Msg("failed to AddTasks")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}

func (s *Server) handleTaskStatistics(w http.ResponseWriter, req *http.Request) {
	response := StatsResponse{}
	waiting, running, err := s.service.GetStatistics()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.log.Err(err).Msg("failed to GetStatistics")
		return
	}
	response.Running = running
	response.Waiting = waiting
	json.NewEncoder(w).Encode(response)
}
