-- xdev module menu / permission / role binding seed for MySQL
-- Execute manually in the admin database.
-- Assumption:
--   - roles PLATFORM_SUPER_ADMIN (tenant_id=0) and SUPER_ADMIN exist
--   - rerunnable via NOT EXISTS guards
--   - sys_apis are synced separately from OpenAPI by admin bootstrap

SET @now = NOW();

-- 1. menus
INSERT INTO sys_menus
(`created_at`,`updated_at`,`created_by`,`updated_by`,`parent_id`,`status`,`type`,`path`,`redirect`,`alias`,`name`,`component`,`meta`)
SELECT
  @now,@now,0,0,NULL,'ON','CATALOG','/xdev','/xdev/device-model-type',NULL,'XdevRoot','BasicLayout',
  JSON_OBJECT(
    'title','设备管理',
    'icon','lucide:cpu',
    'order',90,
    'authority',JSON_ARRAY('xdev:dir')
  )
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menus WHERE parent_id IS NULL AND path = '/xdev'
);

SET @xdev_root_id = (
  SELECT id FROM sys_menus WHERE parent_id IS NULL AND path = '/xdev' LIMIT 1
);

INSERT INTO sys_menus
(`created_at`,`updated_at`,`created_by`,`updated_by`,`parent_id`,`status`,`type`,`path`,`redirect`,`alias`,`name`,`component`,`meta`)
SELECT
  @now,@now,0,0,@xdev_root_id,'ON','MENU','/xdev/device-model-type',NULL,NULL,'XdevDeviceModelType','/xdev/device-model-type/index',
  JSON_OBJECT(
    'title','设备类型',
    'icon','lucide:folder-tree',
    'keepAlive',TRUE,
    'authority',JSON_ARRAY('xdev:device-model-type:view')
  )
FROM DUAL
WHERE @xdev_root_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_menus WHERE parent_id = @xdev_root_id AND path = '/xdev/device-model-type'
  );

INSERT INTO sys_menus
(`created_at`,`updated_at`,`created_by`,`updated_by`,`parent_id`,`status`,`type`,`path`,`redirect`,`alias`,`name`,`component`,`meta`)
SELECT
  @now,@now,0,0,@xdev_root_id,'ON','MENU','/xdev/device-model',NULL,NULL,'XdevDeviceModel','/xdev/device-model/index',
  JSON_OBJECT(
    'title','设备型号',
    'icon','lucide:package-search',
    'keepAlive',TRUE,
    'authority',JSON_ARRAY('xdev:device-model:view')
  )
FROM DUAL
WHERE @xdev_root_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_menus WHERE parent_id = @xdev_root_id AND path = '/xdev/device-model'
  );

INSERT INTO sys_menus
(`created_at`,`updated_at`,`created_by`,`updated_by`,`parent_id`,`status`,`type`,`path`,`redirect`,`alias`,`name`,`component`,`meta`)
SELECT
  @now,@now,0,0,@xdev_root_id,'ON','MENU','/xdev/device',NULL,NULL,'XdevDevice','/xdev/device/index',
  JSON_OBJECT(
    'title','设备信息',
    'icon','lucide:hard-drive',
    'keepAlive',TRUE,
    'authority',JSON_ARRAY('xdev:device:view')
  )
FROM DUAL
WHERE @xdev_root_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_menus WHERE parent_id = @xdev_root_id AND path = '/xdev/device'
  );

-- 2. permissions
INSERT INTO sys_permissions
(`created_at`,`updated_at`,`created_by`,`updated_by`,`status`,`name`,`code`,`group_id`,`description`)
SELECT
  @now,@now,0,0,'ON',seed.name,seed.code,NULL,seed.description
FROM (
  SELECT '设备管理目录' AS name, 'xdev:dir' AS code, '设备管理目录访问权限' AS description
  UNION ALL SELECT '设备类型查看', 'xdev:device-model-type:view', '设备类型页面访问权限'
  UNION ALL SELECT '设备类型新增', 'xdev:device-model-type:create', '设备类型新增权限'
  UNION ALL SELECT '设备类型修改', 'xdev:device-model-type:edit', '设备类型修改权限'
  UNION ALL SELECT '设备类型删除', 'xdev:device-model-type:delete', '设备类型删除权限'
  UNION ALL SELECT '设备类型导出', 'xdev:device-model-type:export', '设备类型导出权限'
  UNION ALL SELECT '设备型号查看', 'xdev:device-model:view', '设备型号页面访问权限'
  UNION ALL SELECT '设备型号新增', 'xdev:device-model:create', '设备型号新增权限'
  UNION ALL SELECT '设备型号修改', 'xdev:device-model:edit', '设备型号修改权限'
  UNION ALL SELECT '设备型号删除', 'xdev:device-model:delete', '设备型号删除权限'
  UNION ALL SELECT '设备型号导出', 'xdev:device-model:export', '设备型号导出权限'
  UNION ALL SELECT '设备信息查看', 'xdev:device:view', '设备信息页面访问权限'
  UNION ALL SELECT '设备信息新增', 'xdev:device:create', '设备信息新增权限'
  UNION ALL SELECT '设备信息修改', 'xdev:device:edit', '设备信息修改权限'
  UNION ALL SELECT '设备信息删除', 'xdev:device:delete', '设备信息删除权限'
  UNION ALL SELECT '设备信息导出', 'xdev:device:export', '设备信息导出权限'
  UNION ALL SELECT '服务查看设备类型', 'service:devicemodeltypeservice:device:view', '设备类型服务查看权限'
  UNION ALL SELECT '服务新增设备类型', 'service:devicemodeltypeservice:device:create', '设备类型服务新增权限'
  UNION ALL SELECT '服务修改设备类型', 'service:devicemodeltypeservice:device:edit', '设备类型服务修改权限'
  UNION ALL SELECT '服务删除设备类型', 'service:devicemodeltypeservice:device:delete', '设备类型服务删除权限'
  UNION ALL SELECT '服务导出设备类型', 'service:devicemodeltypeservice:device:export', '设备类型服务导出权限'
  UNION ALL SELECT '服务查看设备型号', 'service:devicemodelservice:device:view', '设备型号服务查看权限'
  UNION ALL SELECT '服务新增设备型号', 'service:devicemodelservice:device:create', '设备型号服务新增权限'
  UNION ALL SELECT '服务修改设备型号', 'service:devicemodelservice:device:edit', '设备型号服务修改权限'
  UNION ALL SELECT '服务删除设备型号', 'service:devicemodelservice:device:delete', '设备型号服务删除权限'
  UNION ALL SELECT '服务导出设备型号', 'service:devicemodelservice:device:export', '设备型号服务导出权限'
  UNION ALL SELECT '服务查看设备信息', 'service:deviceservice:devices:view', '设备信息服务查看权限'
  UNION ALL SELECT '服务新增设备信息', 'service:deviceservice:devices:create', '设备信息服务新增权限'
  UNION ALL SELECT '服务修改设备信息', 'service:deviceservice:devices:edit', '设备信息服务修改权限'
  UNION ALL SELECT '服务删除设备信息', 'service:deviceservice:devices:delete', '设备信息服务删除权限'
  UNION ALL SELECT '服务导出设备信息', 'service:deviceservice:devices:export', '设备信息服务导出权限'
) AS seed
WHERE NOT EXISTS (
  SELECT 1 FROM sys_permissions p WHERE p.code = seed.code
);

-- 3. permission-menu binding
INSERT INTO sys_permission_menus
(`created_at`,`updated_at`,`created_by`,`updated_by`,`permission_id`,`menu_id`)
SELECT
  @now,@now,0,0,p.id,m.id
FROM (
  SELECT 'xdev:dir' AS code, '/xdev' AS path
  UNION ALL SELECT 'xdev:device-model-type:view', '/xdev/device-model-type'
  UNION ALL SELECT 'xdev:device-model:view', '/xdev/device-model'
  UNION ALL SELECT 'xdev:device:view', '/xdev/device'
) AS seed
JOIN sys_permissions p ON p.code = seed.code
JOIN sys_menus m ON m.path = seed.path
WHERE NOT EXISTS (
  SELECT 1 FROM sys_permission_menus pm
  WHERE pm.permission_id = p.id AND pm.menu_id = m.id
);

-- 4. role-permission binding
INSERT INTO sys_role_permissions
(`created_at`,`updated_at`,`created_by`,`updated_by`,`tenant_id`,`status`,`role_id`,`permission_id`,`effect`,`priority`)
SELECT
  @now,@now,0,0,role_seed.tenant_id,'ON',role_seed.role_id,p.id,'ALLOW',0
FROM (
  SELECT id AS role_id, 0 AS tenant_id
  FROM sys_roles
  WHERE tenant_id = 0 AND code = 'PLATFORM_SUPER_ADMIN'
  LIMIT 1
) AS role_seed
JOIN sys_permissions p
  ON p.code IN (
    'xdev:dir',
    'xdev:device-model-type:view',
    'xdev:device-model-type:create',
    'xdev:device-model-type:edit',
    'xdev:device-model-type:delete',
    'xdev:device-model-type:export',
    'xdev:device-model:view',
    'xdev:device-model:create',
    'xdev:device-model:edit',
    'xdev:device-model:delete',
    'xdev:device-model:export',
    'xdev:device:view',
    'xdev:device:create',
    'xdev:device:edit',
    'xdev:device:delete',
    'xdev:device:export',
    'service:devicemodeltypeservice:device:view',
    'service:devicemodeltypeservice:device:create',
    'service:devicemodeltypeservice:device:edit',
    'service:devicemodeltypeservice:device:delete',
    'service:devicemodeltypeservice:device:export',
    'service:devicemodelservice:device:view',
    'service:devicemodelservice:device:create',
    'service:devicemodelservice:device:edit',
    'service:devicemodelservice:device:delete',
    'service:devicemodelservice:device:export',
    'service:deviceservice:devices:view',
    'service:deviceservice:devices:create',
    'service:deviceservice:devices:edit',
    'service:deviceservice:devices:delete',
    'service:deviceservice:devices:export'
  )
WHERE NOT EXISTS (
  SELECT 1 FROM sys_role_permissions rp
  WHERE rp.role_id = role_seed.role_id AND rp.permission_id = p.id
);

INSERT INTO sys_role_permissions
(`created_at`,`updated_at`,`created_by`,`updated_by`,`tenant_id`,`status`,`role_id`,`permission_id`,`effect`,`priority`)
SELECT
  @now,@now,0,0,role_seed.tenant_id,'ON',role_seed.role_id,p.id,'ALLOW',0
FROM (
  SELECT id AS role_id, COALESCE(tenant_id, 1) AS tenant_id
  FROM sys_roles
  WHERE code = 'SUPER_ADMIN'
  ORDER BY tenant_id ASC, id ASC
  LIMIT 1
) AS role_seed
JOIN sys_permissions p
  ON p.code IN (
    'xdev:dir',
    'xdev:device-model-type:view',
    'xdev:device-model-type:create',
    'xdev:device-model-type:edit',
    'xdev:device-model-type:delete',
    'xdev:device-model-type:export',
    'xdev:device-model:view',
    'xdev:device-model:create',
    'xdev:device-model:edit',
    'xdev:device-model:delete',
    'xdev:device-model:export',
    'xdev:device:view',
    'xdev:device:create',
    'xdev:device:edit',
    'xdev:device:delete',
    'xdev:device:export',
    'service:devicemodeltypeservice:device:view',
    'service:devicemodeltypeservice:device:create',
    'service:devicemodeltypeservice:device:edit',
    'service:devicemodeltypeservice:device:delete',
    'service:devicemodeltypeservice:device:export',
    'service:devicemodelservice:device:view',
    'service:devicemodelservice:device:create',
    'service:devicemodelservice:device:edit',
    'service:devicemodelservice:device:delete',
    'service:devicemodelservice:device:export',
    'service:deviceservice:devices:view',
    'service:deviceservice:devices:create',
    'service:deviceservice:devices:edit',
    'service:deviceservice:devices:delete',
    'service:deviceservice:devices:export'
  )
WHERE NOT EXISTS (
  SELECT 1 FROM sys_role_permissions rp
  WHERE rp.role_id = role_seed.role_id AND rp.permission_id = p.id
);

-- 5. verification helpers
SELECT id, parent_id, path, name, component, meta FROM sys_menus WHERE path LIKE '/xdev%';
SELECT id, code, name, description FROM sys_permissions WHERE code LIKE 'xdev:%' OR code LIKE 'service:%device%';
SELECT role_id, permission_id, tenant_id, status, effect
FROM sys_role_permissions
WHERE permission_id IN (
  SELECT id FROM sys_permissions
  WHERE code LIKE 'xdev:%' OR code LIKE 'service:%device%'
);
