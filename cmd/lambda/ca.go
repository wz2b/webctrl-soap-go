package main

import (
	_ "embed"
)

//go:embed InCommonRSAOVSSLCA3.pem
var inCommonRSAOVSSLCA3 []byte

//go:embed certs/emSignRootTLSCA-G1-corrected-7-8-26.crt
var root1 []byte

//go:embed certs/emSignRootCA-G1.crt
var root2 []byte

//go:embed certs/emSignRootTLSCA-G3.crt
var root3 []byte

// func installExtraCA() error {
//
// 	roots, err := x509.SystemCertPool()
// 	if err != nil {
// 		return fmt.Errorf("load system cert pool: %w", err)
// 	}
// 	if roots == nil {
// 		roots = x509.NewCertPool()
// 	}
//
// 	if ok := roots.AppendCertsFromPEM(inCommonRSAOVSSLCA3); !ok {
// 		return fmt.Errorf("append InCommon RSA OV SSL CA 3 certificate failed")
// 	} else {
//
// 		fmt.Printf("append embedded CA subjects=%d\n", len(roots.Subjects()))
// 	}
// 	http.DefaultTransport = &http.Transport{
// 		TLSClientConfig: &tls.Config{
// 			RootCAs: roots,
// 		},
// 	}
//
// 	return nil
// }
