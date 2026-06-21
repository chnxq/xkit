-- xdev module menu / permission / role binding seed for MySQL
-- Execute manually in the admin database.
-- Assumption:
--   - root catalog is created at top level
--   - roles PLATFORM_SUPER_ADMIN (tenant_id=0) and SUPER_ADMIN exist
--   - rerunnable via NOT EXISTS guards

SET @now = NOW();

-- 1. menus
INSERT INTO sys_menus
(`created_at`,`updated_at`,`created_by`,`updated_by`,`parent_id`,`remark`,`status`,`type`,`path`,`redirect`,`alias`,`name`,`component`,`meta`)
SELECT
  @now,@now,0,0,NULL,NULL,'ON','CATALOG','/xdev','/xdev/device-model-type',NULL,'XdevRoot','BasicLayout',
  JSON_OBJECT(
    'title','menu.xdev.moduleName',
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
(`created_at`,`updated_at`,`created_by`,`updated_by`,`parent_id`,`remark`,`status`,`type`,`path`,`redirect`,`alias`,`name`,`component`,`meta`)
SELECT
  @now,@now,0,0,@xdev_root_id,NULL,'ON','MENU','/xdev/device-model-type',NULL,NULL,'XdevDeviceModelType','/xdev/device-model-type/index',
  JSON_OBJECT(
    'title','menu.xdev.deviceModelType',
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
(`created_at`,`updated_at`,`created_by`,`updated_by`,`parent_id`,`remark`,`status`,`type`,`path`,`redirect`,`alias`,`name`,`component`,`meta`)
SELECT
  @now,@now,0,0,@xdev_root_id,NULL,'ON','MENU','/xdev/device-model',NULL,NULL,'XdevDeviceModel','/xdev/device-model/index',
  JSON_OBJECT(
    'title','menu.xdev.deviceModel',
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
(`created_at`,`updated_at`,`created_by`,`updated_by`,`parent_id`,`remark`,`status`,`type`,`path`,`redirect`,`alias`,`name`,`component`,`meta`)
SELECT
  @now,@now,0,0,@xdev_root_id,NULL,'ON','MENU','/xdev/device',NULL,NULL,'XdevDevice','/xdev/device/index',
  JSON_OBJECT(
    'title','menu.xdev.device',
    'icon','lucide:hard-drive',
    'keepAlive',TRUE,
    'authority',JSON_ARRAY('xdev:device:view')
  )
FROM DUAL
WHERE @xdev_root_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_menus WHERE parent_id = @xdev_root_id AND path = '/xdev/device'
  );

SET @menu_xdev_root = (
  SELECT id FROM sys_menus WHERE parent_id IS NULL AND path = '/xdev' LIMIT 1
);
SET @menu_xdev_device_model_type = (
  SELECT id FROM sys_menus WHERE parent_id = @menu_xdev_root AND path = '/xdev/device-model-type' LIMIT 1
);
SET @menu_xdev_device_model = (
  SELECT id FROM sys_menus WHERE parent_id = @menu_xdev_root AND path = '/xdev/device-model' LIMIT 1
);
SET @menu_xdev_device = (
  SELECT id FROM sys_menus WHERE parent_id = @menu_xdev_root AND path = '/xdev/device' LIMIT 1
);

-- 2. permissions
INSERT INTO sys_permissions
(`created_at`,`updated_at`,`created_by`,`updated_by`,`status`,`name`,`code`,`group_id`,`remark`,`description`)
SELECT
  @now,@now,0,0,'ON','设备管理目录','xdev:dir',NULL,NULL,'xdev module root catalog permission'
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM sys_permissions WHERE code = 'xdev:dir');

INSERT INTO sys_permissions
(`created_at`,`updated_at`,`created_by`,`updated_by`,`status`,`name`,`code`,`group_id`,`remark`,`description`)
SELECT
  @now,@now,0,0,'ON','设备类型查看','xdev:device-model-type:view',NULL,NULL,'xdev device model type page access'
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM sys_permissions WHERE code = 'xdev:device-model-type:view');

INSERT INTO sys_permissions
(`created_at`,`updated_at`,`created_by`,`updated_by`,`status`,`name`,`code`,`group_id`,`remark`,`description`)
SELECT
  @now,@now,0,0,'ON','设备型号查看','xdev:device-model:view',NULL,NULL,'xdev device model page access'
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM sys_permissions WHERE code = 'xdev:device-model:view');

INSERT INTO sys_permissions
(`created_at`,`updated_at`,`created_by`,`updated_by`,`status`,`name`,`code`,`group_id`,`remark`,`description`)
SELECT
  @now,@now,0,0,'ON','设备信息查看','xdev:device:view',NULL,NULL,'xdev device page access'
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM sys_permissions WHERE code = 'xdev:device:view');

SET @perm_xdev_root = (
  SELECT id FROM sys_permissions WHERE code = 'xdev:dir' LIMIT 1
);
SET @perm_xdev_device_model_type = (
  SELECT id FROM sys_permissions WHERE code = 'xdev:device-model-type:view' LIMIT 1
);
SET @perm_xdev_device_model = (
  SELECT id FROM sys_permissions WHERE code = 'xdev:device-model:view' LIMIT 1
);
SET @perm_xdev_device = (
  SELECT id FROM sys_permissions WHERE code = 'xdev:device:view' LIMIT 1
);

-- 3. permission-menu binding
INSERT INTO sys_permission_menus
(`created_at`,`updated_at`,`created_by`,`updated_by`,`permission_id`,`menu_id`)
SELECT @now,@now,0,0,@perm_xdev_root,@menu_xdev_root
FROM DUAL
WHERE @perm_xdev_root IS NOT NULL AND @menu_xdev_root IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_permission_menus
    WHERE permission_id = @perm_xdev_root AND menu_id = @menu_xdev_root
  );

INSERT INTO sys_permission_menus
(`created_at`,`updated_at`,`created_by`,`updated_by`,`permission_id`,`menu_id`)
SELECT @now,@now,0,0,@perm_xdev_device_model_type,@menu_xdev_device_model_type
FROM DUAL
WHERE @perm_xdev_device_model_type IS NOT NULL AND @menu_xdev_device_model_type IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_permission_menus
    WHERE permission_id = @perm_xdev_device_model_type AND menu_id = @menu_xdev_device_model_type
  );

INSERT INTO sys_permission_menus
(`created_at`,`updated_at`,`created_by`,`updated_by`,`permission_id`,`menu_id`)
SELECT @now,@now,0,0,@perm_xdev_device_model,@menu_xdev_device_model
FROM DUAL
WHERE @perm_xdev_device_model IS NOT NULL AND @menu_xdev_device_model IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_permission_menus
    WHERE permission_id = @perm_xdev_device_model AND menu_id = @menu_xdev_device_model
  );

INSERT INTO sys_permission_menus
(`created_at`,`updated_at`,`created_by`,`updated_by`,`permission_id`,`menu_id`)
SELECT @now,@now,0,0,@perm_xdev_device,@menu_xdev_device
FROM DUAL
WHERE @perm_xdev_device IS NOT NULL AND @menu_xdev_device IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_permission_menus
    WHERE permission_id = @perm_xdev_device AND menu_id = @menu_xdev_device
  );

-- 4. role-permission binding
SET @role_platform_super_admin = (
  SELECT id FROM sys_roles WHERE tenant_id = 0 AND code = 'PLATFORM_SUPER_ADMIN' LIMIT 1
);
SET @role_super_admin = (
  SELECT id FROM sys_roles WHERE code = 'SUPER_ADMIN' ORDER BY tenant_id ASC, id ASC LIMIT 1
);
SET @role_super_admin_tenant = (
  SELECT tenant_id FROM sys_roles WHERE id = @role_super_admin LIMIT 1
);

INSERT INTO sys_role_permissions
(`created_at`,`updated_at`,`created_by`,`updated_by`,`tenant_id`,`status`,`role_id`,`permission_id`,`effect`,`priority`)
SELECT @now,@now,0,0,0,'ON',@role_platform_super_admin,@perm_xdev_root,'ALLOW',0
FROM DUAL
WHERE @role_platform_super_admin IS NOT NULL AND @perm_xdev_root IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_role_permissions
    WHERE role_id = @role_platform_super_admin AND permission_id = @perm_xdev_root
  );

INSERT INTO sys_role_permissions
(`created_at`,`updated_at`,`created_by`,`updated_by`,`tenant_id`,`status`,`role_id`,`permission_id`,`effect`,`priority`)
SELECT @now,@now,0,0,0,'ON',@role_platform_super_admin,@perm_xdev_device_model_type,'ALLOW',0
FROM DUAL
WHERE @role_platform_super_admin IS NOT NULL AND @perm_xdev_device_model_type IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_role_permissions
    WHERE role_id = @role_platform_super_admin AND permission_id = @perm_xdev_device_model_type
  );

INSERT INTO sys_role_permissions
(`created_at`,`updated_at`,`created_by`,`updated_by`,`tenant_id`,`status`,`role_id`,`permission_id`,`effect`,`priority`)
SELECT @now,@now,0,0,0,'ON',@role_platform_super_admin,@perm_xdev_device_model,'ALLOW',0
FROM DUAL
WHERE @role_platform_super_admin IS NOT NULL AND @perm_xdev_device_model IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_role_permissions
    WHERE role_id = @role_platform_super_admin AND permission_id = @perm_xdev_device_model
  );

INSERT INTO sys_role_permissions
(`created_at`,`updated_at`,`created_by`,`updated_by`,`tenant_id`,`status`,`role_id`,`permission_id`,`effect`,`priority`)
SELECT @now,@now,0,0,0,'ON',@role_platform_super_admin,@perm_xdev_device,'ALLOW',0
FROM DUAL
WHERE @role_platform_super_admin IS NOT NULL AND @perm_xdev_device IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_role_permissions
    WHERE role_id = @role_platform_super_admin AND permission_id = @perm_xdev_device
  );

INSERT INTO sys_role_permissions
(`created_at`,`updated_at`,`created_by`,`updated_by`,`tenant_id`,`status`,`role_id`,`permission_id`,`effect`,`priority`)
SELECT @now,@now,0,0,COALESCE(@role_super_admin_tenant,1),'ON',@role_super_admin,@perm_xdev_root,'ALLOW',0
FROM DUAL
WHERE @role_super_admin IS NOT NULL AND @perm_xdev_root IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_role_permissions
    WHERE role_id = @role_super_admin AND permission_id = @perm_xdev_root
  );

INSERT INTO sys_role_permissions
(`created_at`,`updated_at`,`created_by`,`updated_by`,`tenant_id`,`status`,`role_id`,`permission_id`,`effect`,`priority`)
SELECT @now,@now,0,0,COALESCE(@role_super_admin_tenant,1),'ON',@role_super_admin,@perm_xdev_device_model_type,'ALLOW',0
FROM DUAL
WHERE @role_super_admin IS NOT NULL AND @perm_xdev_device_model_type IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_role_permissions
    WHERE role_id = @role_super_admin AND permission_id = @perm_xdev_device_model_type
  );

INSERT INTO sys_role_permissions
(`created_at`,`updated_at`,`created_by`,`updated_by`,`tenant_id`,`status`,`role_id`,`permission_id`,`effect`,`priority`)
SELECT @now,@now,0,0,COALESCE(@role_super_admin_tenant,1),'ON',@role_super_admin,@perm_xdev_device_model,'ALLOW',0
FROM DUAL
WHERE @role_super_admin IS NOT NULL AND @perm_xdev_device_model IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_role_permissions
    WHERE role_id = @role_super_admin AND permission_id = @perm_xdev_device_model
  );

INSERT INTO sys_role_permissions
(`created_at`,`updated_at`,`created_by`,`updated_by`,`tenant_id`,`status`,`role_id`,`permission_id`,`effect`,`priority`)
SELECT @now,@now,0,0,COALESCE(@role_super_admin_tenant,1),'ON',@role_super_admin,@perm_xdev_device,'ALLOW',0
FROM DUAL
WHERE @role_super_admin IS NOT NULL AND @perm_xdev_device IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM sys_role_permissions
    WHERE role_id = @role_super_admin AND permission_id = @perm_xdev_device
  );

-- 5. verification helpers
SELECT id, parent_id, path, name, component FROM sys_menus WHERE path LIKE '/xdev%';
SELECT id, code, name FROM sys_permissions WHERE code LIKE 'xdev:%';
SELECT role_id, permission_id, tenant_id, status, effect
FROM sys_role_permissions
WHERE permission_id IN (@perm_xdev_root, @perm_xdev_device_model_type, @perm_xdev_device_model, @perm_xdev_device);
