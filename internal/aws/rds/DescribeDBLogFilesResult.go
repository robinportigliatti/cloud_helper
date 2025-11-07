package rds

type DescribeDBLogFilesResult struct {
	DescribeDBLogFiles []struct {
		LogFileName string `json:"LogFileName"`
		LastWritten int64  `json:"LastWritten"`
		Size        int    `json:"Size"`
	} `json:"DescribeDBLogFiles"`
}
