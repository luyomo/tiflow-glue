use sequence_safe_mode_test;
alter table t2 add column age2 int after age;
alter table t2 add index age2(age2);
insert into t2 (uid, name, age, age2) values (40004, 'Remedios Moscote', 100, 300), (40005, 'Amaranta', 103, 301);
insert into t3 (uid, name, age) values (30007, 'Aureliano José', 99), (30008, 'Santa Sofía de la Piedad', 999), (30009, '17 Aurelianos', 9999);
update t2 set age = age + 33 where uid = 40004;
update t3 set age = age + 44 where uid > 30006 and uid < 30010;
alter table t3 add column age2 int after age;
alter table t3 add index age2(age2);
update t3 set age2 = 100;
