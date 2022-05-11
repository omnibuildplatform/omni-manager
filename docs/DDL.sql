CREATE DATABASE  IF NOT EXISTS `obs_meta` /*!40100 DEFAULT CHARACTER SET utf8mb4 */;
USE `obs_meta`;
-- MySQL dump 10.13  Distrib 8.0.28, for Win64 (x86_64)
--
-- Host: 192.168.1.193    Database: obs_meta
-- ------------------------------------------------------
-- Server version	5.5.5-10.7.3-MariaDB-1:10.7.3+maria~focal

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `image_meta`
--

DROP TABLE IF EXISTS `image_meta`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `image_meta` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `packages` varchar(100) DEFAULT NULL COMMENT 'architecture',
  `version` varchar(45) DEFAULT NULL COMMENT 'release openEuler Version',
  `build_type` varchar(50) DEFAULT NULL COMMENT 'iso , zip ....',
  `base_pkg` mediumtext DEFAULT NULL COMMENT '默认的基本package',
  `custom_pkg` mediumtext DEFAULT NULL COMMENT '自定义',
  `create_time` datetime DEFAULT NULL COMMENT 'create time',
  `job_name` varchar(150) DEFAULT NULL,
  `status` varchar(10) NOT NULL DEFAULT 'running' COMMENT 'current status :1 :submit request   2 build   3finished',
  `user_id` int(11) DEFAULT NULL COMMENT 'user id',
  `user_name` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `query_index` (`packages`,`version`,`build_type`),
  KEY `userid_index` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='meta data for image ';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2022-03-21 16:11:46
