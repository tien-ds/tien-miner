package ip2region

var reIp *Ip2Region

func GetIp() *Ip2Region {
	if reIp != nil {
		return reIp
	}
	region, err := New()
	if err != nil {
		return nil
	}
	reIp = region
	return reIp
}
