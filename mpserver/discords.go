package main

import (
	"errors"
	"strconv"
	"time"

	"github.com/aouyang1/go-matrixprofile/matrixprofile"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

type Discord struct {
	Groups []int       `json:"groups"`
	Series [][]float64 `json:"series"`
}

func topKDiscords(c *gin.Context) {
	start := time.Now()
	endpoint := "/api/v1/topkdiscords"
	method := "GET"
	session := sessions.Default(c)
	buildCORSHeaders(c)

	kstr := c.Query("k")

	k, err := strconv.Atoi(kstr)
	if err != nil {
		requestTotal.WithLabelValues(method, endpoint, "500").Inc()
		serviceRequestDuration.WithLabelValues(endpoint).Observe(time.Since(start).Seconds() * 1000)
		glog.Infof("%v", err)
		c.JSON(500, RespError{Error: err.Error()})
		return
	}

	v := fetchMPCache(session)
	var mp matrixprofile.MatrixProfile
	if v == nil {
		requestTotal.WithLabelValues(method, endpoint, "500").Inc()
		serviceRequestDuration.WithLabelValues(endpoint).Observe(time.Since(start).Seconds() * 1000)
		err := errors.New("matrix profile is not initialized to compute discords")
		glog.Infof("%v", err)
		c.JSON(500, RespError{err.Error(), true})
		return
	}
	mp = v.(matrixprofile.MatrixProfile)
	discords, err := mp.TopKDiscords(k, mp.M/2)
	if err != nil {
		requestTotal.WithLabelValues(method, endpoint, "500").Inc()
		serviceRequestDuration.WithLabelValues(endpoint).Observe(time.Since(start).Seconds() * 1000)
		err := errors.New("failed to compute discords")
		glog.Infof("%v", err)
		c.JSON(500, RespError{Error: err.Error()})
		return
	}

	var discord Discord
	discord.Groups = discords
	discord.Series = make([][]float64, len(discords))
	for i, didx := range discord.Groups {
		discord.Series[i], err = matrixprofile.ZNormalize(mp.A[didx : didx+mp.M])
		if err != nil {
			requestTotal.WithLabelValues(method, endpoint, "500").Inc()
			serviceRequestDuration.WithLabelValues(endpoint).Observe(time.Since(start).Seconds() * 1000)
			glog.Infof("%v", err)
			c.JSON(500, RespError{Error: err.Error()})
			return
		}
	}

	requestTotal.WithLabelValues(method, endpoint, "200").Inc()
	serviceRequestDuration.WithLabelValues(endpoint).Observe(time.Since(start).Seconds() * 1000)
	c.JSON(200, discord)
}
