use sequence_safe_mode_test;
insert into t1 (uid, name, age) values (10009, 'Buenos Aires', 100);
insert into t2 (uid, name, age) values (20008, 'Colonel Aureliano Buendía', 402);
alter table t1 add column age2 int after age;
alter table t1 add index age2(age2);
insert into t1 (uid, name, age, age2) values (10010, 'Buenos Aires', 200, 404);
insert into t2 (uid, name, age) values (20009, 'Colonel Aureliano Buendía', 100);
update t2 set age = age + 1 where uid = 20008;
update t2 set age = age + 2 where uid = 20007;
update t2 set age = age + 1 where uid = 20009;
update t1 set age = age + 1 where uid = 10008;
update t1 set age = age + 1 where uid = 10009;
alter table t2 add column age2 int after age;
alter table t2 add index age2(age2);
update t1 set age = age + 1 where uid in (10010, 10011);
update t2 set age = age + 1 where uid in (20009, 20010);
