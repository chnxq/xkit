-- xdev module menu / permission / role binding seed for PostgreSQL
-- Execute manually in the admin database.
-- Assumption:
--   - root catalog is created at top level
--   - roles PLATFORM_SUPER_ADMIN (tenant_id=0) and SUPER_ADMIN exist
--   - rerunnable via NOT EXISTS guards

DO $$
DECLARE
  v_now timestamp := NOW();
  v_xdev_root_id integer;
  v_menu_xdev_root integer;
  v_menu_xdev_device_model_type integer;
  v_menu_xdev_device_model integer;
  v_menu_xdev_device integer;
  v_perm_xdev_root integer;
  v_perm_xdev_device_model_type integer;
  v_perm_xdev_device_model integer;
  v_perm_xdev_device integer;
  v_role_platform_super_admin integer;
  v_role_super_admin integer;
  v_role_super_admin_tenant integer;
BEGIN
  INSERT INTO sys_menus
  (created_at,updated_at,created_by,updated_by,parent_id,status,type,path,redirect,alias,name,component,meta)
  SELECT
    v_now,v_now,0,0,NULL,'ON','CATALOG','/xdev','/xdev/device-model-type',NULL,'XdevRoot','BasicLayout',
    jsonb_build_object(
      'title','menu.xdev.moduleName',
      'icon','lucide:cpu',
      'order',90,
      'authority',jsonb_build_array('xdev:dir')
    )
  WHERE NOT EXISTS (
    SELECT 1 FROM sys_menus WHERE parent_id IS NULL AND path = '/xdev'
  );

  SELECT id INTO v_xdev_root_id
  FROM sys_menus
  WHERE parent_id IS NULL AND path = '/xdev'
  LIMIT 1;

  INSERT INTO sys_menus
  (created_at,updated_at,created_by,updated_by,parent_id,status,type,path,redirect,alias,name,component,meta)
  SELECT
    v_now,v_now,0,0,v_xdev_root_id,'ON','MENU','/xdev/device-model-type',NULL,NULL,'XdevDeviceModelType','/xdev/device-model-type/index',
    jsonb_build_object(
      'title','menu.xdev.deviceModelType',
      'icon','lucide:folder-tree',
      'keepAlive',true,
      'authority',jsonb_build_array('xdev:device-model-type:view')
    )
  WHERE v_xdev_root_id IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM sys_menus WHERE parent_id = v_xdev_root_id AND path = '/xdev/device-model-type'
    );

  INSERT INTO sys_menus
  (created_at,updated_at,created_by,updated_by,parent_id,status,type,path,redirect,alias,name,component,meta)
  SELECT
    v_now,v_now,0,0,v_xdev_root_id,'ON','MENU','/xdev/device-model',NULL,NULL,'XdevDeviceModel','/xdev/device-model/index',
    jsonb_build_object(
      'title','menu.xdev.deviceModel',
      'icon','lucide:package-search',
      'keepAlive',true,
      'authority',jsonb_build_array('xdev:device-model:view')
    )
  WHERE v_xdev_root_id IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM sys_menus WHERE parent_id = v_xdev_root_id AND path = '/xdev/device-model'
    );

  INSERT INTO sys_menus
  (created_at,updated_at,created_by,updated_by,parent_id,status,type,path,redirect,alias,name,component,meta)
  SELECT
    v_now,v_now,0,0,v_xdev_root_id,'ON','MENU','/xdev/device',NULL,NULL,'XdevDevice','/xdev/device/index',
    jsonb_build_object(
      'title','menu.xdev.device',
      'icon','lucide:hard-drive',
      'keepAlive',true,
      'authority',jsonb_build_array('xdev:device:view')
    )
  WHERE v_xdev_root_id IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM sys_menus WHERE parent_id = v_xdev_root_id AND path = '/xdev/device'
    );

  SELECT id INTO v_menu_xdev_root
  FROM sys_menus
  WHERE parent_id IS NULL AND path = '/xdev'
  LIMIT 1;

  SELECT id INTO v_menu_xdev_device_model_type
  FROM sys_menus
  WHERE parent_id = v_menu_xdev_root AND path = '/xdev/device-model-type'
  LIMIT 1;

  SELECT id INTO v_menu_xdev_device_model
  FROM sys_menus
  WHERE parent_id = v_menu_xdev_root AND path = '/xdev/device-model'
  LIMIT 1;

  SELECT id INTO v_menu_xdev_device
  FROM sys_menus
  WHERE parent_id = v_menu_xdev_root AND path = '/xdev/device'
  LIMIT 1;

  INSERT INTO sys_permissions
  (created_at,updated_at,created_by,updated_by,status,name,code,group_id,description)
  SELECT
    v_now,v_now,0,0,'ON','设备管理目录','xdev:dir',NULL,'xdev module root catalog permission'
  WHERE NOT EXISTS (SELECT 1 FROM sys_permissions WHERE code = 'xdev:dir');

  INSERT INTO sys_permissions
  (created_at,updated_at,created_by,updated_by,status,name,code,group_id,description)
  SELECT
    v_now,v_now,0,0,'ON','设备类型查看','xdev:device-model-type:view',NULL,'xdev device model type page access'
  WHERE NOT EXISTS (SELECT 1 FROM sys_permissions WHERE code = 'xdev:device-model-type:view');

  INSERT INTO sys_permissions
  (created_at,updated_at,created_by,updated_by,status,name,code,group_id,description)
  SELECT
    v_now,v_now,0,0,'ON','设备型号查看','xdev:device-model:view',NULL,'xdev device model page access'
  WHERE NOT EXISTS (SELECT 1 FROM sys_permissions WHERE code = 'xdev:device-model:view');

  INSERT INTO sys_permissions
  (created_at,updated_at,created_by,updated_by,status,name,code,group_id,description)
  SELECT
    v_now,v_now,0,0,'ON','设备信息查看','xdev:device:view',NULL,'xdev device page access'
  WHERE NOT EXISTS (SELECT 1 FROM sys_permissions WHERE code = 'xdev:device:view');

  SELECT id INTO v_perm_xdev_root FROM sys_permissions WHERE code = 'xdev:dir' LIMIT 1;
  SELECT id INTO v_perm_xdev_device_model_type FROM sys_permissions WHERE code = 'xdev:device-model-type:view' LIMIT 1;
  SELECT id INTO v_perm_xdev_device_model FROM sys_permissions WHERE code = 'xdev:device-model:view' LIMIT 1;
  SELECT id INTO v_perm_xdev_device FROM sys_permissions WHERE code = 'xdev:device:view' LIMIT 1;

  INSERT INTO sys_permission_menus
  (created_at,updated_at,created_by,updated_by,permission_id,menu_id)
  SELECT v_now,v_now,0,0,v_perm_xdev_root,v_menu_xdev_root
  WHERE v_perm_xdev_root IS NOT NULL AND v_menu_xdev_root IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM sys_permission_menus
      WHERE permission_id = v_perm_xdev_root AND menu_id = v_menu_xdev_root
    );

  INSERT INTO sys_permission_menus
  (created_at,updated_at,created_by,updated_by,permission_id,menu_id)
  SELECT v_now,v_now,0,0,v_perm_xdev_device_model_type,v_menu_xdev_device_model_type
  WHERE v_perm_xdev_device_model_type IS NOT NULL AND v_menu_xdev_device_model_type IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM sys_permission_menus
      WHERE permission_id = v_perm_xdev_device_model_type AND menu_id = v_menu_xdev_device_model_type
    );

  INSERT INTO sys_permission_menus
  (created_at,updated_at,created_by,updated_by,permission_id,menu_id)
  SELECT v_now,v_now,0,0,v_perm_xdev_device_model,v_menu_xdev_device_model
  WHERE v_perm_xdev_device_model IS NOT NULL AND v_menu_xdev_device_model IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM sys_permission_menus
      WHERE permission_id = v_perm_xdev_device_model AND menu_id = v_menu_xdev_device_model
    );

  INSERT INTO sys_permission_menus
  (created_at,updated_at,created_by,updated_by,permission_id,menu_id)
  SELECT v_now,v_now,0,0,v_perm_xdev_device,v_menu_xdev_device
  WHERE v_perm_xdev_device IS NOT NULL AND v_menu_xdev_device IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM sys_permission_menus
      WHERE permission_id = v_perm_xdev_device AND menu_id = v_menu_xdev_device
    );

  SELECT id INTO v_role_platform_super_admin
  FROM sys_roles
  WHERE tenant_id = 0 AND code = 'PLATFORM_SUPER_ADMIN'
  LIMIT 1;

  SELECT id, tenant_id INTO v_role_super_admin, v_role_super_admin_tenant
  FROM sys_roles
  WHERE code = 'SUPER_ADMIN'
  ORDER BY tenant_id ASC, id ASC
  LIMIT 1;

  INSERT INTO sys_role_permissions
  (created_at,updated_at,created_by,updated_by,tenant_id,status,role_id,permission_id,effect,priority)
  SELECT v_now,v_now,0,0,0,'ON',v_role_platform_super_admin,v_perm_xdev_root,'ALLOW',0
  WHERE v_role_platform_super_admin IS NOT NULL AND v_perm_xdev_root IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM sys_role_permissions
      WHERE role_id = v_role_platform_super_admin AND permission_id = v_perm_xdev_root
    );

  INSERT INTO sys_role_permissions
  (created_at,updated_at,created_by,updated_by,tenant_id,status,role_id,permission_id,effect,priority)
  SELECT v_now,v_now,0,0,0,'ON',v_role_platform_super_admin,v_perm_xdev_device_model_type,'ALLOW',0
  WHERE v_role_platform_super_admin IS NOT NULL AND v_perm_xdev_device_model_type IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM sys_role_permissions
      WHERE role_id = v_role_platform_super_admin AND permission_id = v_perm_xdev_device_model_type
    );

  INSERT INTO sys_role_permissions
  (created_at,updated_at,created_by,updated_by,tenant_id,status,role_id,permission_id,effect,priority)
  SELECT v_now,v_now,0,0,0,'ON',v_role_platform_super_admin,v_perm_xdev_device_model,'ALLOW',0
  WHERE v_role_platform_super_admin IS NOT NULL AND v_perm_xdev_device_model IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM sys_role_permissions
      WHERE role_id = v_role_platform_super_admin AND permission_id = v_perm_xdev_device_model
    );

  INSERT INTO sys_role_permissions
  (created_at,updated_at,created_by,updated_by,tenant_id,status,role_id,permission_id,effect,priority)
  SELECT v_now,v_now,0,0,0,'ON',v_role_platform_super_admin,v_perm_xdev_device,'ALLOW',0
  WHERE v_role_platform_super_admin IS NOT NULL AND v_perm_xdev_device IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM sys_role_permissions
      WHERE role_id = v_role_platform_super_admin AND permission_id = v_perm_xdev_device
    );

  INSERT INTO sys_role_permissions
  (created_at,updated_at,created_by,updated_by,tenant_id,status,role_id,permission_id,effect,priority)
  SELECT v_now,v_now,0,0,COALESCE(v_role_super_admin_tenant,1),'ON',v_role_super_admin,v_perm_xdev_root,'ALLOW',0
  WHERE v_role_super_admin IS NOT NULL AND v_perm_xdev_root IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM sys_role_permissions
      WHERE role_id = v_role_super_admin AND permission_id = v_perm_xdev_root
    );

  INSERT INTO sys_role_permissions
  (created_at,updated_at,created_by,updated_by,tenant_id,status,role_id,permission_id,effect,priority)
  SELECT v_now,v_now,0,0,COALESCE(v_role_super_admin_tenant,1),'ON',v_role_super_admin,v_perm_xdev_device_model_type,'ALLOW',0
  WHERE v_role_super_admin IS NOT NULL AND v_perm_xdev_device_model_type IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM sys_role_permissions
      WHERE role_id = v_role_super_admin AND permission_id = v_perm_xdev_device_model_type
    );

  INSERT INTO sys_role_permissions
  (created_at,updated_at,created_by,updated_by,tenant_id,status,role_id,permission_id,effect,priority)
  SELECT v_now,v_now,0,0,COALESCE(v_role_super_admin_tenant,1),'ON',v_role_super_admin,v_perm_xdev_device_model,'ALLOW',0
  WHERE v_role_super_admin IS NOT NULL AND v_perm_xdev_device_model IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM sys_role_permissions
      WHERE role_id = v_role_super_admin AND permission_id = v_perm_xdev_device_model
    );

  INSERT INTO sys_role_permissions
  (created_at,updated_at,created_by,updated_by,tenant_id,status,role_id,permission_id,effect,priority)
  SELECT v_now,v_now,0,0,COALESCE(v_role_super_admin_tenant,1),'ON',v_role_super_admin,v_perm_xdev_device,'ALLOW',0
  WHERE v_role_super_admin IS NOT NULL AND v_perm_xdev_device IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM sys_role_permissions
      WHERE role_id = v_role_super_admin AND permission_id = v_perm_xdev_device
    );
END
$$;

SELECT id, parent_id, path, name, component FROM sys_menus WHERE path LIKE '/xdev%';
SELECT id, code, name FROM sys_permissions WHERE code LIKE 'xdev:%';
SELECT role_id, permission_id, tenant_id, status, effect
FROM sys_role_permissions
WHERE permission_id IN (
  (SELECT id FROM sys_permissions WHERE code = 'xdev:dir' LIMIT 1),
  (SELECT id FROM sys_permissions WHERE code = 'xdev:device-model-type:view' LIMIT 1),
  (SELECT id FROM sys_permissions WHERE code = 'xdev:device-model:view' LIMIT 1),
  (SELECT id FROM sys_permissions WHERE code = 'xdev:device:view' LIMIT 1)
);
