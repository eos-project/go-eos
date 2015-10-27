package server

import "github.com/eos-project/go-eos/model"

type Listener func(model.Message)
