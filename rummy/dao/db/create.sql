CREATE TABLE `order_202403`
(
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT "自增ID",
    `room_id` int NOT NULL DEFAULT 0 COMMENT "房间ID",
    `player_id` bigint NOT NULL DEFAULT 0 COMMENT "用户ID",
    `order_no` varchar(32) NOT NULL DEFAULT "" COMMENT "订单号",
    `round_no` varchar(20) NOT NULL DEFAULT "" COMMENT "期号",
    `item_id` bigint NOT NULL DEFAULT 0 COMMENT "注点",
    `bet_chips` bigint NOT NULL DEFAULT 0 COMMENT "下注大小",
    `currency` varchar(10) NOT NULL DEFAULT "" COMMENT "货币",
    `prize` bigint NOT NULL DEFAULT 0 COMMENT "中奖额度",
    `timestamp` bigint NOT NULL DEFAULT 0 COMMENT "时间戳",
    `status` tinyint NOT NULL DEFAULT 0 COMMENT "订单状态 0 下单成功 1结算成功 2结算失败",
    PRIMARY KEY (`id`),
    KEY (`room_id`,`round_no`,`player_id`),
    KEY (`player_id`),
    UNIQUE (`order_no`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 AUTO_INCREMENT=1000;

CREATE TABLE `order_202404` LIKE `order_202403`;
CREATE TABLE `order_202405` LIKE `order_202403`;
CREATE TABLE `order_202406` LIKE `order_202403`;
CREATE TABLE `order_202407` LIKE `order_202403`;
CREATE TABLE `order_202408` LIKE `order_202403`;
CREATE TABLE `order_202409` LIKE `order_202403`;
CREATE TABLE `order_202410` LIKE `order_202403`;
CREATE TABLE `order_202411` LIKE `order_202403`;
CREATE TABLE `order_202412` LIKE `order_202403`;


CREATE TABLE `history_202403`
(
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT "自增ID",
    `room_id` int NOT NULL DEFAULT 0 COMMENT "房间ID",
    `player_id` bigint NOT NULL DEFAULT 0 COMMENT "用户ID",
    `round_no` varchar(32) NOT NULL DEFAULT "" COMMENT "期号",
    `begin_time` bigint NOT NULL DEFAULT 0 COMMENT "本期开始时间戳",
    `end_time` bigint NOT NULL DEFAULT 0 COMMENT "本期结束时间戳",
    `item_id` bigint NOT NULL DEFAULT 0 COMMENT "注点",
    `total_bet` bigint NOT NULL DEFAULT 0 COMMENT "总下注",
    `total_prize` bigint NOT NULL DEFAULT 0 COMMENT "总奖励",
    PRIMARY KEY (`id`),
    UNIQUE (`room_id`, `round_no`, `player_id`, `item_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 AUTO_INCREMENT=1000;

CREATE TABLE `history_202404` LIKE `history_202403`;
CREATE TABLE `history_202405` LIKE `history_202403`;
CREATE TABLE `history_202406` LIKE `history_202403`;
CREATE TABLE `history_202407` LIKE `history_202403`;
CREATE TABLE `history_202408` LIKE `history_202403`;
CREATE TABLE `history_202409` LIKE `history_202403`;
CREATE TABLE `history_202410` LIKE `history_202403`;
CREATE TABLE `history_202411` LIKE `history_202403`;
CREATE TABLE `history_202412` LIKE `history_202403`;

CREATE TABLE `lottery_result`
(
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT "自增ID",
    `room_id` int NOT NULL DEFAULT 0 COMMENT "房间ID",
    `round_no` varchar(32) NOT NULL DEFAULT "" COMMENT "期号",
    `result_mode` tinyint NOT NULL DEFAULT 0 COMMENT "0表示随机 1表示杀率控制",
    `winner` tinyint NOT NULL DEFAULT 0 COMMENT "1:红 2:黑",
    `red_cards` varchar(20) NOT NULL DEFAULT "" COMMENT "红方牌",
    `red_type` int NOT NULL DEFAULT 0 COMMENT "获胜方牌型 11:高牌 12:小对 13:大对 14:同花 15:顺子 16:同花顺 17:三条",
    `black_cards` varchar(20) NOT NULL DEFAULT "" COMMENT "黑方牌",
    `black_type` int NOT NULL DEFAULT 0 COMMENT "获胜方牌型 11:高牌 12:小对 13:大对 14:同花 15:顺子 16:同花顺 17:三条",
    `total_bet` bigint NOT NULL DEFAULT 0 COMMENT "总下注",
    `total_prize` bigint NOT NULL DEFAULT 0 COMMENT "总奖励",
    `settle_timestamp` bigint NOT NULL DEFAULT 0 COMMENT "开奖时间戳",
    PRIMARY KEY (`id`),
    UNIQUE (`room_id`, `round_no`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 AUTO_INCREMENT=1000;

CREATE TABLE `stock`
(
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT "自增ID",
    `room_id` int NOT NULL DEFAULT 0 COMMENT "房间ID",
    `number` bigint NOT NULL DEFAULT 0 COMMENT "库存数量",
    PRIMARY KEY (`id`),
    UNIQUE (`room_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 AUTO_INCREMENT=1000;
