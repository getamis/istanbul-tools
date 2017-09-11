// Copyright 2017 AMIS Technologies
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package log

import (
	"os"

	"github.com/inconshreveable/log15"
)

var defaultLogger = log15.New()

func init() {
	defaultLogger.SetHandler(
		log15.MultiHandler(
			log15.CallerFileHandler(log15.StreamHandler(
				os.Stdout, log15.TerminalFormat(),
			)),
		),
	)
}

func New(ctx ...interface{}) log15.Logger {
	return defaultLogger.New(ctx...)
}

func Info(msg string, ctx ...interface{}) {
	defaultLogger.Info(msg, ctx...)
}

func Debug(msg string, ctx ...interface{}) {
	defaultLogger.Debug(msg, ctx...)
}

func Warn(msg string, ctx ...interface{}) {
	defaultLogger.Warn(msg, ctx)
}

func Error(msg string, ctx ...interface{}) {
	defaultLogger.Error(msg, ctx...)
}

func Fatal(msg string, ctx ...interface{}) {
	defaultLogger.Crit(msg, ctx...)
}
