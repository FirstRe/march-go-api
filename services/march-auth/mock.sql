

INSERT INTO functions (id, name) VALUES
  ('c1b57d27-5d3f-4496-a18e-a621e63d358f', 'MENU:CUSTOMER'),
  ('c99bc7f8-a272-4813-be1b-16ede3146438', 'MENU:DASHBOARD'),
  ('96a74331-8713-4c0b-b558-f0dcd6186066', 'MENU:INVENTORY'),
  ('5d510dc7-3302-4d19-b2ef-b97a19139e16', 'MENU:SALES'),
  ('3787f89a-4791-408e-83da-df4d7bf732f7', 'MENU:USER');

INSERT INTO shops (id, name, description, created_by, updated_by, created_at, updated_at)
VALUES
	('984d0d87-7d74-45c5-9d94-6ebcb74a98de','first_shop','test_auto_create','system','system','2024-11-26 08:34:11.672','2024-11-26 08:34:11.672');


INSERT INTO groups (id, name, shops_id)
VALUES
	('9596c980-1cae-4bac-b27b-b54bbac9bffb','SuperAdmin|first_shop','984d0d87-7d74-45c5-9d94-6ebcb74a98de');

INSERT INTO tasks (id, name, function_id, description)
VALUES
	('033c261a-b4fd-466b-9f17-8dd011abff5e','INBrandMaker','96a74331-8713-4c0b-b558-f0dcd6186066','จัดการยี่ห้อสินค้า|Brand Maker'),
	('08747b9f-6ebd-469f-aa58-df35788c52c9','INTypeMaker','96a74331-8713-4c0b-b558-f0dcd6186066','จัดการประเภทสินค้า|Type Maker'),
	('2720bbfb-8824-45c7-8261-770892a88ccb','INTrashMaker','96a74331-8713-4c0b-b558-f0dcd6186066','จัดการถังขยะ|Trash Maker'),
	('349221fc-8939-4c4f-aa95-3d99d713cc42','INBrandViewer','96a74331-8713-4c0b-b558-f0dcd6186066','ดูยี่ห้อสินค้า|Brand Viewer'),
	('499f1e4b-90bc-4d17-9286-104ce53c5e88','INTypeViewer','96a74331-8713-4c0b-b558-f0dcd6186066','ดูประเภทสินค้า|Type Viewer'),
	('5e2adbb0-d8aa-4744-8006-584bf810e352','INViewer','96a74331-8713-4c0b-b558-f0dcd6186066','ดูสินค้า|Inventory Viewer'),
	('5fe1d67e-e52b-494e-bccf-709059ff297c','INBranchMaker','96a74331-8713-4c0b-b558-f0dcd6186066','จัดการสาขาสินค้า|Branch Maker'),
	('945abef9-5ea9-447b-8a7e-aa3b22ea63d2','INCSV','96a74331-8713-4c0b-b558-f0dcd6186066','จัดการ Csv|Csv Maker'),
	('991b691c-0fe4-47e4-a39d-a8d04390cff7','INBranchViewer','96a74331-8713-4c0b-b558-f0dcd6186066','ดูสาขาสินค้า|Branch Viewer'),
	('c9610668-5cd0-415e-a0b7-decd78f7a494','INMaker','96a74331-8713-4c0b-b558-f0dcd6186066','จัดการสินค้า|Inventory Maker');

INSERT INTO group_tasks (id, name, group_id, task_id, shops_id, created_by, updated_by, created_at, updated_at)
VALUES
	('24140f79-2bcc-4860-9da2-46d86e605bad','SuperAdmin|first_shop_INBranchViewer','9596c980-1cae-4bac-b27b-b54bbac9bffb','991b691c-0fe4-47e4-a39d-a8d04390cff7','984d0d87-7d74-45c5-9d94-6ebcb74a98de','system','system','2024-11-26 08:34:11.702','2024-11-26 08:34:11.702'),
	('2662e6b0-540e-47c4-b687-e95bea6e8c2e','SuperAdmin|first_shop_INBranchMaker','9596c980-1cae-4bac-b27b-b54bbac9bffb','5fe1d67e-e52b-494e-bccf-709059ff297c','984d0d87-7d74-45c5-9d94-6ebcb74a98de','system','system','2024-11-26 08:34:11.702','2024-11-26 08:34:11.702'),
	('344aca71-53b9-4e11-ab2f-4d07a8af368c','SuperAdmin|first_shop_INBrandViewer','9596c980-1cae-4bac-b27b-b54bbac9bffb','349221fc-8939-4c4f-aa95-3d99d713cc42','984d0d87-7d74-45c5-9d94-6ebcb74a98de','system','system','2024-11-26 08:34:11.702','2024-11-26 08:34:11.702'),
	('47bbeed0-46f7-40b2-b225-83d261811640','SuperAdmin|first_shop_INMaker','9596c980-1cae-4bac-b27b-b54bbac9bffb','c9610668-5cd0-415e-a0b7-decd78f7a494','984d0d87-7d74-45c5-9d94-6ebcb74a98de','system','system','2024-11-26 08:34:11.702','2024-11-26 08:34:11.702'),
	('a3792cf1-2e34-4143-a7a9-4d006d9eaa01','SuperAdmin|first_shop_INBrandMaker','9596c980-1cae-4bac-b27b-b54bbac9bffb','033c261a-b4fd-466b-9f17-8dd011abff5e','984d0d87-7d74-45c5-9d94-6ebcb74a98de','system','system','2024-11-26 08:34:11.702','2024-11-26 08:34:11.702'),
	('b8943dec-db9c-456d-be34-9c374bab0d38','SuperAdmin|first_shop_INViewer','9596c980-1cae-4bac-b27b-b54bbac9bffb','5e2adbb0-d8aa-4744-8006-584bf810e352','984d0d87-7d74-45c5-9d94-6ebcb74a98de','system','system','2024-11-26 08:34:11.702','2024-11-26 08:34:11.702'),
	('b98a57b7-3cb4-4cad-8501-9d3e5e8599ba','SuperAdmin|first_shop_INCSV','9596c980-1cae-4bac-b27b-b54bbac9bffb','945abef9-5ea9-447b-8a7e-aa3b22ea63d2','984d0d87-7d74-45c5-9d94-6ebcb74a98de','system','system','2024-11-26 08:34:11.702','2024-11-26 08:34:11.702'),
	('ebd57550-9639-484b-b34c-6a23631e8ba9','SuperAdmin|first_shop_INTrashMaker','9596c980-1cae-4bac-b27b-b54bbac9bffb','2720bbfb-8824-45c7-8261-770892a88ccb','984d0d87-7d74-45c5-9d94-6ebcb74a98de','system','system','2024-11-26 08:34:11.702','2024-11-26 08:34:11.702'),
	('ef7cd58f-b04f-4b09-b70e-17947b5cd9f5','SuperAdmin|first_shop_INTypeMaker','9596c980-1cae-4bac-b27b-b54bbac9bffb','08747b9f-6ebd-469f-aa58-df35788c52c9','984d0d87-7d74-45c5-9d94-6ebcb74a98de','system','system','2024-11-26 08:34:11.702','2024-11-26 08:34:11.702'),
	('f810b54d-52f1-4c72-9549-dd9e0a619e3b','SuperAdmin|first_shop_INTypeViewer','9596c980-1cae-4bac-b27b-b54bbac9bffb','499f1e4b-90bc-4d17-9286-104ce53c5e88','984d0d87-7d74-45c5-9d94-6ebcb74a98de','system','system','2024-11-26 08:34:11.702','2024-11-26 08:34:11.702');


INSERT INTO group_functions (id, name, function_id, group_id, "create", view, "update")
VALUES
	('1d9869e1-5cc4-4d9f-9b52-e95a657b61ec','SuperAdmin|first_shop_MENU:INVENTORY','96a74331-8713-4c0b-b558-f0dcd6186066','9596c980-1cae-4bac-b27b-b54bbac9bffb',TRUE,TRUE,TRUE),
	('30280acc-9311-4ecb-9e2e-66cf3faf4e91','SuperAdmin|first_shop_MENU:USER','3787f89a-4791-408e-83da-df4d7bf732f7','9596c980-1cae-4bac-b27b-b54bbac9bffb',TRUE,TRUE,TRUE),
	('3501fc95-a3c0-4809-b7f4-485ef05177fc','SuperAdmin|first_shop_MENU:DASHBOARD','c99bc7f8-a272-4813-be1b-16ede3146438','9596c980-1cae-4bac-b27b-b54bbac9bffb',TRUE,TRUE,TRUE),
	('3d29c52e-ce71-4780-a8de-343da699c0ac','SuperAdmin|first_shop_MENU:CUSTOMER','c1b57d27-5d3f-4496-a18e-a621e63d358f','9596c980-1cae-4bac-b27b-b54bbac9bffb',TRUE,TRUE,TRUE),
	('72957b42-7b40-4a7a-82ce-1218bae99b8e','SuperAdmin|first_shop_MENU:SALES','5d510dc7-3302-4d19-b2ef-b97a19139e16','9596c980-1cae-4bac-b27b-b54bbac9bffb',TRUE,TRUE,TRUE);


INSERT INTO users (id, group_id, shops_id, username, password, is_super_admin, email, is_registered, picture, refresh_token, deleted, device_id, created_by, updated_by, created_at, updated_at)
VALUES
	('f8bab2a4-cb39-499c-827c-c22507369847','9596c980-1cae-4bac-b27b-b54bbac9bffb','984d0d87-7d74-45c5-9d94-6ebcb74a98de','NochTich',NULL,TRUE,'firstzaxshot95@gmail.com',TRUE,'https://lh3.googleusercontent.com/a/ACg8ocJkBRKtkxE8xXWMctj0qiWn0m0-JqXSBXk1hXAw_qHnMRPyRi66=s96-c','eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImY4YmFiMmE0LWNiMzktNDk5Yy04MjdjLWMyMjUwNzM2OTg0NyIsImRldmljZUlkIjoiYjc5ZDBiZTYtNTQ3Yi00MWQzLTljMDEtZGIwYjYyZTViMTgyIiwiZXhwIjoxOTg0NjMyMzk0LCJpYXQiOjE3NDUxMzE1OTR9.WiMj_9XOfP60y14uhlrjtDm3LZK5dsWtQr7okF4uio4',FALSE,'b79d0be6-547b-41d3-9c01-db0b62e5b182','system','system','2024-11-26 08:34:11.768','2025-04-20 13:46:34.795');
