-- Description: MySQL test data for xdev current schema
-- Scope:
--   1. xdev module tables
--   2. minimal host user/org-unit helper rows for relation display
-- Target tenants:
--   - tenant_id = 0 (platform/global sample)
--   - tenant_id = 1 (tenant sample)

SET FOREIGN_KEY_CHECKS = 0;
START TRANSACTION;

-- ---------------------------------------------------------------------------
-- Minimal host helper data for relation display in xdev UI
-- ---------------------------------------------------------------------------

DELETE FROM sys_user_org_units WHERE user_id BETWEEN 9001 AND 9010;
DELETE FROM sys_users WHERE id BETWEEN 9001 AND 9010;
DELETE FROM sys_org_units WHERE id BETWEEN 9101 AND 9110;

INSERT INTO sys_org_units (
  id, tenant_id, parent_id, type, name, code, description, path, sort_order, status, created_at, updated_at
)
VALUES
  (9101, 0, NULL, 'COMPANY', '平台设备运营中心', 'PLT-DEV-OPS', '平台级设备运营组织', '/9101', 1, 'ON', NOW(), NOW()),
  (9102, 0, 9101, 'DEPARTMENT', '平台网络资源组', 'PLT-NET', '平台网络设备组织', '/9101/9102', 2, 'ON', NOW(), NOW()),
  (9111, 1, NULL, 'COMPANY', '默认租户设备部', 'TENANT1-DEV', '租户一设备组织', '/9111', 1, 'ON', NOW(), NOW()),
  (9112, 1, 9111, 'DEPARTMENT', '默认租户网络组', 'TENANT1-NET', '租户一网络设备组织', '/9111/9112', 2, 'ON', NOW(), NOW());

INSERT INTO sys_users (
  id, tenant_id, username, nickname, realname, email, mobile, status, gender, created_at, updated_at
)
VALUES
  (9001, 0, 'xdev_platform_ops', '平台运维', '平台运维负责人', 'xdev_platform_ops@example.local', '13900009001', 'NORMAL', 'MALE', NOW(), NOW()),
  (9002, 0, 'xdev_platform_net', '平台网络', '平台网络管理员', 'xdev_platform_net@example.local', '13900009002', 'NORMAL', 'FEMALE', NOW(), NOW()),
  (9011, 1, 'xdev_tenant_admin', '租户设备', '租户设备管理员', 'xdev_tenant_admin@example.local', '13900009111', 'NORMAL', 'MALE', NOW(), NOW()),
  (9012, 1, 'xdev_tenant_net', '租户网络', '租户网络管理员', 'xdev_tenant_net@example.local', '13900009112', 'NORMAL', 'FEMALE', NOW(), NOW());

INSERT INTO sys_user_org_units (user_id, org_unit_id)
VALUES
  (9001, 9101),
  (9002, 9102),
  (9011, 9111),
  (9012, 9112);

-- ---------------------------------------------------------------------------
-- xdev cleanup
-- ---------------------------------------------------------------------------

DELETE FROM xdev_dev_group_device WHERE id BETWEEN 8601 AND 8799;
DELETE FROM xdev_dev_group_user WHERE id BETWEEN 8801 AND 8899;
DELETE FROM xdev_dev_group_org_unit WHERE id BETWEEN 8901 AND 8999;
DELETE FROM xdev_dev_info WHERE id BETWEEN 8301 AND 8499;
DELETE FROM xdev_dev_group WHERE id BETWEEN 8501 AND 8599;
DELETE FROM xdev_dev_model WHERE id BETWEEN 8201 AND 8299;
DELETE FROM xdev_dev_model_type WHERE id BETWEEN 8101 AND 8199;

-- ---------------------------------------------------------------------------
-- xdev model types
-- ---------------------------------------------------------------------------

INSERT INTO xdev_dev_model_type (
  id, created_by, created_at, updated_at, deleted_at, tenant_id,
  model_type_name, use_case, type_desc
)
VALUES
  (8101, 1, NOW(), NOW(), NULL, 0, '平台网络设备', 'NETWORK', '平台统一管理的网络基础设备'),
  (8102, 1, NOW(), NOW(), NULL, 0, '平台安防设备', 'SECURITY', '平台统一管理的安防类设备'),
  (8111, 2, NOW(), NOW(), NULL, 1, '租户办公设备', 'OFFICE', '租户日常办公终端与打印设备'),
  (8112, 2, NOW(), NOW(), NULL, 1, '租户生产设备', 'PRODUCTION', '租户业务侧生产采集与控制设备');

-- ---------------------------------------------------------------------------
-- xdev models
-- ---------------------------------------------------------------------------

INSERT INTO xdev_dev_model (
  id, created_by, created_at, updated_at, deleted_at, tenant_id,
  model_name, description, remark, model_type_id
)
VALUES
  (8201, 1, NOW(), NOW(), NULL, 0, 'SW-AGG-48', '48口平台汇聚交换机', '平台网络主力型号', 8101),
  (8202, 1, NOW(), NOW(), NULL, 0, 'CAM-PTZ-4K', '4K云台安防摄像机', '平台园区视频型号', 8102),
  (8211, 2, NOW(), NOW(), NULL, 1, 'NB-DEV-14', '14寸租户办公笔记本', '默认办公终端', 8111),
  (8212, 2, NOW(), NOW(), NULL, 1, 'PRT-LASER-A4', 'A4激光打印机', '行政共享打印', 8111),
  (8213, 2, NOW(), NOW(), NULL, 1, 'PLC-CTRL-X1', '轻量控制器', '租户生产控制器', 8112),
  (8214, 2, NOW(), NOW(), NULL, 1, 'SENSOR-TEMP-X', '温湿度采集终端', '租户环境采集终端', 8112);

-- ---------------------------------------------------------------------------
-- xdev devices
-- ---------------------------------------------------------------------------

INSERT INTO xdev_dev_info (
  id, created_by, updated_by, created_at, updated_at, deleted_at, tenant_id,
  device_code, name, serial_number, finger_print, use_status, meta_data, model_id
)
VALUES
  (8301, 1, 1, NOW(), NOW(), NULL, 0, 'PLT-SW-001', '平台核心交换机-A', 'SN-PLT-SW-001', NULL, 'USING', NULL, 8201),
  (8302, 1, 1, NOW(), NOW(), NULL, 0, 'PLT-SW-002', '平台核心交换机-B', 'SN-PLT-SW-002', NULL, 'IDLE', NULL, 8201),
  (8303, 1, 1, NOW(), NOW(), NULL, 0, 'PLT-CAM-001', '平台园区摄像头-东门', 'SN-PLT-CAM-001', NULL, 'USING', NULL, 8202),
  (8304, 1, 1, NOW(), NOW(), NULL, 0, 'PLT-CAM-002', '平台园区摄像头-西门', 'SN-PLT-CAM-002', NULL, 'REPAIR', NULL, 8202),
  (8311, 2, 2, NOW(), NOW(), NULL, 1, 'TEN-NB-001', '租户办公笔记本-张三', 'SN-TEN-NB-001', NULL, 'USING', NULL, 8211),
  (8312, 2, 2, NOW(), NOW(), NULL, 1, 'TEN-NB-002', '租户办公笔记本-李四', 'SN-TEN-NB-002', NULL, 'IDLE', NULL, 8211),
  (8313, 2, 2, NOW(), NOW(), NULL, 1, 'TEN-PRT-001', '租户共享打印机-一楼', 'SN-TEN-PRT-001', NULL, 'USING', NULL, 8212),
  (8314, 2, 2, NOW(), NOW(), NULL, 1, 'TEN-PLC-001', '租户生产控制器-1号线', 'SN-TEN-PLC-001', NULL, 'USING', NULL, 8213),
  (8315, 2, 2, NOW(), NOW(), NULL, 1, 'TEN-PLC-002', '租户生产控制器-2号线', 'SN-TEN-PLC-002', NULL, 'DISABLED', NULL, 8213),
  (8316, 2, 2, NOW(), NOW(), NULL, 1, 'TEN-SENSOR-001', '租户环境传感器-仓库A', 'SN-TEN-SENSOR-001', NULL, 'USING', NULL, 8214),
  (8317, 2, 2, NOW(), NOW(), NULL, 1, 'TEN-SENSOR-002', '租户环境传感器-仓库B', 'SN-TEN-SENSOR-002', NULL, 'SCRAPPED', NULL, 8214);

-- ---------------------------------------------------------------------------
-- xdev groups (tree)
-- ---------------------------------------------------------------------------

INSERT INTO xdev_dev_group (
  id, created_by, updated_by, created_at, updated_at, deleted_at, tenant_id,
  status, sort_order, path, group_name, type, is_leaf_node, descript, visible, parent_id
)
VALUES
  (8501, 1, 1, NOW(), NOW(), NULL, 0, 'ON', 1, '/8501/', '平台设备中心', 'FUNCTION', 0, '平台总设备目录', 1, NULL),
  (8502, 1, 1, NOW(), NOW(), NULL, 0, 'ON', 1, '/8501/8502/', '平台网络组', 'NETWORK', 1, '平台网络设备叶子组', 1, 8501),
  (8503, 1, 1, NOW(), NOW(), NULL, 0, 'ON', 2, '/8501/8503/', '平台安防组', 'FUNCTION', 1, '平台安防设备叶子组', 1, 8501),
  (8511, 2, 2, NOW(), NOW(), NULL, 1, 'ON', 1, '/8511/', '租户设备中心', 'FUNCTION', 0, '租户总设备目录', 1, NULL),
  (8512, 2, 2, NOW(), NOW(), NULL, 1, 'ON', 1, '/8511/8512/', '租户办公组', 'DEPARTMENT', 1, '租户办公设备叶子组', 1, 8511),
  (8513, 2, 2, NOW(), NOW(), NULL, 1, 'ON', 2, '/8511/8513/', '租户生产组', 'FUNCTION', 1, '租户生产设备叶子组', 1, 8511),
  (8514, 2, 2, NOW(), NOW(), NULL, 1, 'OFF', 3, '/8511/8514/', '租户环境监测组', 'NETWORK', 1, '租户环境设备叶子组', 0, 8511);

-- ---------------------------------------------------------------------------
-- xdev group-device relations
-- ---------------------------------------------------------------------------

INSERT INTO xdev_dev_group_device (
  id, created_by, created_at, updated_at, deleted_at, tenant_id, device_id, group_id
)
VALUES
  (8601, 1, NOW(), NOW(), NULL, 0, 8301, 8502),
  (8602, 1, NOW(), NOW(), NULL, 0, 8302, 8502),
  (8603, 1, NOW(), NOW(), NULL, 0, 8303, 8503),
  (8604, 1, NOW(), NOW(), NULL, 0, 8304, 8503),
  (8611, 2, NOW(), NOW(), NULL, 1, 8311, 8512),
  (8612, 2, NOW(), NOW(), NULL, 1, 8312, 8512),
  (8613, 2, NOW(), NOW(), NULL, 1, 8313, 8512),
  (8614, 2, NOW(), NOW(), NULL, 1, 8314, 8513),
  (8615, 2, NOW(), NOW(), NULL, 1, 8315, 8513),
  (8616, 2, NOW(), NOW(), NULL, 1, 8316, 8514),
  (8617, 2, NOW(), NOW(), NULL, 1, 8317, 8514);

-- ---------------------------------------------------------------------------
-- xdev group-user relations
-- ---------------------------------------------------------------------------

INSERT INTO xdev_dev_group_user (
  id, created_by, created_at, updated_at, deleted_at, tenant_id, user_id, group_id
)
VALUES
  (8801, 1, NOW(), NOW(), NULL, 0, 9001, 8502),
  (8802, 1, NOW(), NOW(), NULL, 0, 9002, 8503),
  (8811, 2, NOW(), NOW(), NULL, 1, 9011, 8512),
  (8812, 2, NOW(), NOW(), NULL, 1, 9012, 8513);

-- ---------------------------------------------------------------------------
-- xdev group-org-unit relations
-- ---------------------------------------------------------------------------

INSERT INTO xdev_dev_group_org_unit (
  id, created_by, created_at, updated_at, deleted_at, tenant_id, org_unit_id, group_id
)
VALUES
  (8901, 1, NOW(), NOW(), NULL, 0, 9102, 8502),
  (8902, 1, NOW(), NOW(), NULL, 0, 9101, 8503),
  (8911, 2, NOW(), NOW(), NULL, 1, 9111, 8512),
  (8912, 2, NOW(), NOW(), NULL, 1, 9112, 8513),
  (8913, 2, NOW(), NOW(), NULL, 1, 9112, 8514);

ALTER TABLE xdev_dev_model_type AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM xdev_dev_model_type);
ALTER TABLE xdev_dev_model AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM xdev_dev_model);
ALTER TABLE xdev_dev_info AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM xdev_dev_info);
ALTER TABLE xdev_dev_group AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM xdev_dev_group);
ALTER TABLE xdev_dev_group_device AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM xdev_dev_group_device);
ALTER TABLE xdev_dev_group_user AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM xdev_dev_group_user);
ALTER TABLE xdev_dev_group_org_unit AUTO_INCREMENT = (SELECT COALESCE(MAX(id) + 1, 1) FROM xdev_dev_group_org_unit);

COMMIT;
SET FOREIGN_KEY_CHECKS = 1;
