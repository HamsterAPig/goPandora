package model

type PandoraParam struct {
	ApiPrefix             string
	PandoraSentry         bool
	BuildId               string
	EnableSharePageVerify bool
}

var Param PandoraParam
