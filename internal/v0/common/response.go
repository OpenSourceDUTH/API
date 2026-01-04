package common

import (
	"time"

	"github.com/google/uuid"
)

// Structs for the API response format

type Metadata struct {
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	RequestID string    `json:"requestId"`
}

type APIResponse struct {
	Data     interface{} `json:"data"`
	Errors   []string    `json:"errors"`
	Metadata Metadata    `json:"metadata"`
}

// Response functions

func CreateAPIResponse(data interface{}, errors []string, requestID string) APIResponse {
	// If the requestID is blank and not cascading from other functions generate a new one
	if requestID == "" {
		requestID = uuid.New().String()
	}
	return APIResponse{
		Data:   data,
		Errors: errors,
		Metadata: Metadata{
			Timestamp: time.Now(),
			Version:   "v0",
			RequestID: requestID,
		},
	}
}

func CreateSuccessResponse(data interface{}) APIResponse {
	return CreateAPIResponse(
		data,
		[]string{},
		"",
	)
}

func CreateErrorResponse(errors []string) APIResponse {
	return CreateAPIResponse(
		nil,
		errors,
		"",
	)
}

func CreateSuccessResponseWithRequestID(data interface{}, requestID string) APIResponse {
	return CreateAPIResponse(
		data,
		[]string{},
		requestID,
	)
}

func CreateErrorResponseWithRequestID(errors []string, requestID string) APIResponse {
	return CreateAPIResponse(
		nil,
		errors,
		requestID,
	)
}

//This project is the monolithic backend API for the OpenSourceDUTH team. Access to open data compiled and provided by the OpenSourceDUTH University Team.
//API Copyright (C) 2025 OpenSourceDUTH
//This program is free software: you can redistribute it and/or modify
//it under the terms of the GNU General Public License as published by
//the Free Software Foundation, either version 3 of the License, or
//(at your option) any later version.
//
//This program is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//GNU General Public License for more details.
//
//You should have received a copy of the GNU General Public License
//along with this program.  If not, see <https://www.gnu.org/licenses/>.
