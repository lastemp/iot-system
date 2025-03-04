CREATE TABLE `iot_messages` (
  `id` int NOT NULL AUTO_INCREMENT,
  `topic` varchar(300) DEFAULT '',
  `payload` varchar(50) DEFAULT '',
  `date_added` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb3;
