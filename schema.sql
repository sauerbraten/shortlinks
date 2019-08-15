drop table if exists `links`;

create table `links` (
	`id` integer primary key,
	`longURL` text not null unique
);

pragma journal_mode = wal;