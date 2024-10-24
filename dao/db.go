package dao

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func getMysqlCursor(host string, port int, username string, passwd string, dbname string) *sqlx.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=latin1", username, passwd, host, port, dbname)
	db := sqlx.MustConnect("mysql", dsn)

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(4)
	return db
}

var MysqlCursor = getMysqlCursor("127.0.0.1", 2881, "lcx", "11223344", "go-pubchem")

/* default table structure to store pubchem-data
CREATE TABLE `compound_from_pubchem` (
  `cid` int(11) NOT NULL,
  `mw` float DEFAULT NULL,
  `polararea` int(11) DEFAULT NULL,
  `complexity` float DEFAULT NULL,
  `xlogp` float DEFAULT NULL,
  `exactmass` float DEFAULT NULL,
  `monoisotopicmass` float DEFAULT NULL,
  `heavycnt` int(11) DEFAULT NULL,
  `hbonddonor` int(11) DEFAULT NULL,
  `hbondacc` int(11) DEFAULT NULL,
  `rotbonds` int(11) DEFAULT NULL,
  `annothitcnt` int(11) DEFAULT NULL,
  `charge` int(11) DEFAULT NULL,
  `covalentunitcnt` int(11) DEFAULT NULL,
  `isotopeatomcnt` int(11) DEFAULT NULL,
  `totalatomstereocnt` int(11) DEFAULT NULL,
  `definedatomstereocnt` int(11) DEFAULT NULL,
  `undefinedatomstereocnt` int(11) DEFAULT NULL,
  `totalbondstereocnt` int(11) DEFAULT NULL,
  `definedbondstereocnt` int(11) DEFAULT NULL,
  `undefinedbondstereocnt` int(11) DEFAULT NULL,
  `pclidcnt` int(11) DEFAULT NULL,
  `gpidcnt` int(11) DEFAULT NULL,
  `gpfamilycnt` int(11) DEFAULT NULL,
  `aids` varchar(1000) DEFAULT NULL,
  `cmpdname` varchar(1000) DEFAULT NULL,
  `cmpdsynonym` text DEFAULT NULL,
  `inchi` text DEFAULT NULL,
  `inchikey` varchar(1000) DEFAULT NULL,
  `isosmiles` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL,
  `iupacname` varchar(3000) DEFAULT NULL,
  `mf` varchar(1000) DEFAULT NULL,
  `sidsrcname` text DEFAULT NULL,
  `cidcdate` varchar(1000) DEFAULT NULL,
  `depcatg` varchar(1000) DEFAULT NULL,
  `annothits` varchar(1000) DEFAULT NULL,
  `neighbortype` varchar(1000) DEFAULT NULL,
  `canonicalsmiles` varchar(1000) DEFAULT NULL,
  PRIMARY KEY (`cid`)
) DEFAULT CHARSET = utf8mb4
*/
