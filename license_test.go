// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package main

import "testing"
import mt "github.com/mendersoftware/mendertesting"

func TestMenderCompliance(t *testing.T) {
	mt.SetFirstEnterpriseCommit("f9d59fe4eb6be7f57d53a4b76bbf09e487bd00c6")
	mt.CheckMenderCompliance(t)
}
