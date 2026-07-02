<!-- Hand-written by Codex. -->
<script lang="ts" setup>
import type {
  FormInstance,
  TableColumnsType,
  TablePaginationConfig,
  TreeProps,
  TreeSelectProps,
} from 'ant-design-vue';
import type { Key } from 'ant-design-vue/es/_util/type';
import type { DataNode } from 'ant-design-vue/es/tree';

import type { VbenFormProps } from '@vben/common-ui';

import type {
  AdminDevice,
  AdminDeviceListParams,
  AdminDeviceListResult,
} from '../../provider/device.provider';
import type { AdminDeviceGroup } from '../../provider/device-group.provider';
import type {
  AdminDeviceGroupDevice,
  AdminDeviceGroupDeviceSaveInput,
} from '../../provider/device-group-device-rel.provider';
import type {
  AdminDeviceGroupOrgUnit,
  AdminOrgUnitOption,
} from '../../provider/device-group-org-unit.provider';
import type {
  AdminDeviceGroupUser,
  AdminUserOption,
} from '../../provider/device-group-user.provider';
import type { VxeTableGridOptions } from '#/adapter/vxe-table';

import { computed, nextTick, onMounted, reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';
import { useUserStore } from '@vben/stores';

import {
  Button,
  Form,
  Input,
  message,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tree,
} from 'ant-design-vue';
import dayjs from 'dayjs';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { listAdminOrgUnitsApi } from '#/api/admin/org-units';
import { listAdminUsersApi } from '#/api/admin/users';
import AdminGeneratedForm from '#/components/admin-generated-form/index.vue';
import {
  normalizeAdminTableSortDirection,
  normalizeAdminTableSortField,
} from '#/components/admin-table-toolbar/shared';
import { $t } from '#/locales';

import {
  buildFormOptions as buildDeviceFormOptions,
  buildListGridColumns as buildDeviceGridColumns,
} from '../../meta/device.meta';
import { buildFormOptions as buildDeviceGroupFormOptions } from '../../meta/device-group.meta';
import {
  createDevice,
  deleteDevice,
  getDeviceById,
  listDeviceModelOptions,
  listDevicePage,
  updateDevice,
} from '../../provider/device.provider';
import {
  createDeviceGroup,
  deleteDeviceGroup,
  getDeviceGroupById,
  listDeviceGroupPage,
  updateDeviceGroup,
} from '../../provider/device-group.provider';
import {
  createDeviceGroupDevice,
  deleteDeviceGroupDevice,
  listDeviceGroupDevicePage,
} from '../../provider/device-group-device-rel.provider';
import {
  createDeviceGroupOrgUnit,
  deleteDeviceGroupOrgUnit,
  listDeviceGroupOrgUnitPage,
} from '../../provider/device-group-org-unit.provider';
import {
  createDeviceGroupUser,
  deleteDeviceGroupUser,
  listDeviceGroupUserPage,
} from '../../provider/device-group-user.provider';

type DeviceRow = AdminDevice & {
  relationId?: number;
};

type BindDeviceRow = AdminDevice & {
  alreadyBound?: boolean;
};

type GroupOrgUnitRow = {
  id: number;
  name: string;
  relationId?: number;
};

type GroupUserRow = {
  id: number;
  label: string;
  orgUnitNames: string[];
  relationId?: number;
};

type TreeOption = NonNullable<TreeSelectProps['treeData']>[number];

const deviceGroupLoading = ref(false);
const modelLoading = ref(false);
const bindLoading = ref(false);
const groupSubmitting = ref(false);
const deviceSubmitting = ref(false);
const bindingSubmitting = ref(false);
const groupOrgUnitSubmitting = ref(false);
const groupUserSubmitting = ref(false);
const groupModalOpen = ref(false);
const deviceModalOpen = ref(false);
const bindModalOpen = ref(false);
const groupOrgUnitModalOpen = ref(false);
const groupUserModalOpen = ref(false);
const editingGroupId = ref<number>();
const editingDeviceId = ref<number>();
const editingDeviceReadonly = ref(false);
const selectedGroupId = ref<number>();
const selectedGroupTreeKeys = computed(() =>
  selectedGroupId.value ? [String(selectedGroupId.value)] : [],
);
const groupFormRef = ref<FormInstance>();
const deviceFormRef = ref<FormInstance>();
const groupFormModel = reactive<Record<string, any>>({});
const deviceFormModel = reactive<Record<string, any>>({});
const groupTreeItems = ref<AdminDeviceGroup[]>([]);
const deviceModelOptions = ref<Array<{ label: string; value: number | string }>>([]);
const deviceModelDialogOptions = ref<Array<{ label: string; value: number | string }>>([]);
const currentGroupRelationItems = ref<AdminDeviceGroupDevice[]>([]);
const currentGroupOrgUnitItems = ref<AdminDeviceGroupOrgUnit[]>([]);
const currentGroupUserItems = ref<AdminDeviceGroupUser[]>([]);
const boundDeviceIdList = ref<number[]>([]);
const expandedGroupKeys = ref<Array<number | string>>([]);
const groupOrgUnitRows = ref<GroupOrgUnitRow[]>([]);
const groupUserRows = ref<GroupUserRow[]>([]);
const orgUnitDialogOptions = ref<AdminOrgUnitOption[]>([]);
const userDialogOptions = ref<AdminUserOption[]>([]);
const userFilterOrgUnitId = ref<number>();
const selectedOrgUnitIDs = ref<number[]>([]);
const selectedUserIDs = ref<number[]>([]);
const userStore = useUserStore();
const currentTenantId = computed(() => userStore.userInfo?.tenantId ?? 0);

const deviceSearchForm = reactive<AdminDeviceListParams>({
  deviceCode: '',
  modelId: undefined,
  name: '',
  useStatus: undefined,
});

const bindSearchForm = reactive<AdminDeviceListParams>({
  deviceCode: '',
  modelId: undefined,
  name: '',
  useStatus: undefined,
});

const bindPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});
const bindSorting = ref<AdminDeviceListParams['sorting']>(undefined);

const bindTableRows = ref<BindDeviceRow[]>([]);
const bindSelectedRowKeys = ref<number[]>([]);
const bindRowSelection = computed(() => ({
  getCheckboxProps: (record: BindDeviceRow) => ({
    disabled: !!record.alreadyBound,
  }),
  onChange: (keys: (number | string)[]) => {
    bindSelectedRowKeys.value = keys
      .map((key) => Number(key))
      .filter((key) => Number.isFinite(key));
  },
  selectedRowKeys: bindSelectedRowKeys.value,
}));

const deviceModelOptionsMap = computed(
  () => new Map(deviceModelOptions.value.map((item) => [item.value, item.label])),
);

const deviceUseStatusOptions = computed(() => [
  { label: $t('enum.device.useStatus.DISABLED'), value: 'DISABLED' },
  { label: $t('enum.device.useStatus.IDLE'), value: 'IDLE' },
  { label: $t('enum.device.useStatus.REPAIR'), value: 'REPAIR' },
  { label: $t('enum.device.useStatus.SCRAPPED'), value: 'SCRAPPED' },
  { label: $t('enum.device.useStatus.USING'), value: 'USING' },
]);

const deviceUseStatusTextMap = computed(() => ({
  DISABLED: $t('enum.device.useStatus.DISABLED'),
  IDLE: $t('enum.device.useStatus.IDLE'),
  REPAIR: $t('enum.device.useStatus.REPAIR'),
  SCRAPPED: $t('enum.device.useStatus.SCRAPPED'),
  USING: $t('enum.device.useStatus.USING'),
}));

const deviceTypeTextMap = computed(() => ({
  DEPARTMENT: $t('enum.deviceGroup.type.DEPARTMENT'),
  FUNCTION: $t('enum.deviceGroup.type.FUNCTION'),
  NETWORK: $t('enum.deviceGroup.type.NETWORK'),
  USER: $t('enum.deviceGroup.type.USER'),
}));

const selectedGroup = computed(() =>
  findGroupById(groupTreeItems.value, selectedGroupId.value),
);

const selectedGroupLabel = computed(() => {
  if (!selectedGroup.value) {
    return $t('page.deviceGroup.moduleName');
  }
  return selectedGroup.value.groupName ?? `#${selectedGroup.value.id}`;
});

const selectedGroupType = computed(() =>
  resolveGroupTypeText(selectedGroup.value?.type),
);

const selectedGroupIsLeaf = computed(() => !!selectedGroup.value?.isLeafNode);

const selectedGroupBoundCount = computed(() => currentGroupRelationItems.value.length);
const selectedGroupAllKeys = computed(() => getGroupTreeKeys(groupTreeItems.value));
const selectedGroupMutable = computed(() =>
  selectedGroup.value ? canMutateTenant(selectedGroup.value.tenantId) : false,
);
const deviceModalTitle = computed(() =>
  editingDeviceId.value
    ? editingDeviceReadonly.value
      ? $t('common.detail')
      : $t('page.device.editTitle')
    : $t('page.device.createTitle'),
);

const orgUnitTreeOptions = computed<DataNode[]>(() =>
  buildOrgUnitTreeOptions(orgUnitDialogOptions.value),
);

const bindTableColumns: TableColumnsType<BindDeviceRow> = [
  {
    dataIndex: 'deviceCode',
    key: 'deviceCode',
    sorter: true,
    title: $t('page.device.deviceCode'),
    width: 180,
  },
  {
    dataIndex: 'name',
    key: 'name',
    sorter: true,
    title: $t('page.device.name'),
    width: 180,
  },
  {
    dataIndex: 'modelId',
    key: 'modelId',
    title: $t('page.device.modelId'),
    width: 180,
  },
  {
    dataIndex: 'serialNumber',
    key: 'serialNumber',
    sorter: true,
    title: $t('page.device.serialNumber'),
    width: 180,
  },
  {
    dataIndex: 'useStatus',
    key: 'useStatus',
    title: $t('page.device.useStatus'),
    width: 120,
  },
  {
    dataIndex: 'createdAt',
    key: 'createdAt',
    sorter: true,
    title: $t('page.device.createdAt'),
    width: 170,
  },
];

const deviceGridColumns = [
  ...((buildDeviceGridColumns($t) ?? []).filter((column: any) => column.field !== 'action')),
  {
    field: 'action',
    fixed: 'right' as const,
    slots: { default: 'action' },
    title: $t('ui.table.action'),
    width: 180,
  },
];

const deviceGridOptions: VxeTableGridOptions<DeviceRow> = {
  border: false,
  columnConfig: {
    resizable: true,
  },
  columns: deviceGridColumns,
  exportConfig: {
    filename: 'xdev-device-center',
    type: 'csv',
  },
  height: 'auto',
  keepSource: true,
  pagerConfig: {},
  proxyConfig: {
    ajax: {
      query: async (
        { page, sort }: { page: { currentPage: number; pageSize: number }; sort: { field?: string; order?: string } },
      ) => {
        return await loadDevicePage({
          page: page?.currentPage ?? 1,
          pageSize: page?.pageSize ?? 20,
          sorting:
            sort?.field && normalizeAdminTableSortDirection(sort.order)
              ? [
                  {
                    direction: normalizeAdminTableSortDirection(sort.order)!,
                    field: toBackendSortField(String(sort.field)),
                  },
                ]
              : undefined,
        });
      },
    },
    sort: true,
  },
  rowConfig: {
    isHover: true,
  },
  stripe: true,
  toolbarConfig: {
    custom: true,
    export: true,
    refresh: true,
    slots: {
      toolPrefix: 'toolPrefix',
    },
    zoom: true,
  },
};

const groupFormOptions = computed<VbenFormProps>(() => {
  const base = buildDeviceGroupFormOptions($t);
  return {
    ...base,
    schema: (base.schema || []).map((item: any) => {
      switch (item.fieldName) {
        case 'parentId':
          return {
            ...item,
            component: 'TreeSelect',
            componentProps: {
              ...(resolveComponentProps(item) || {}),
              allowClear: true,
              placeholder:
                $t('page.deviceGroup.selectParentId') ||
                resolveComponentProps(item).placeholder,
              showSearch: true,
              treeData: buildParentTreeOptions(groupTreeItems.value, editingGroupId.value),
              treeDefaultExpandAll: true,
              treeNodeFilterProp: 'title',
            },
          };
        case 'type':
          return {
            ...item,
            componentProps: {
              ...(resolveComponentProps(item) || {}),
              allowClear: true,
              options: [
                { label: $t('enum.deviceGroup.type.DEPARTMENT'), value: 'DEPARTMENT' },
                { label: $t('enum.deviceGroup.type.FUNCTION'), value: 'FUNCTION' },
                { label: $t('enum.deviceGroup.type.NETWORK'), value: 'NETWORK' },
                { label: $t('enum.deviceGroup.type.USER'), value: 'USER' },
              ],
              placeholder: $t('page.deviceGroup.selectType'),
            },
          };
        case 'status':
          return {
            ...item,
            componentProps: {
              ...(resolveComponentProps(item) || {}),
              allowClear: true,
              options: [
                { label: $t('enum.deviceGroup.status.ON'), value: 'ON' },
                { label: $t('enum.deviceGroup.status.OFF'), value: 'OFF' },
              ],
            },
          };
        default:
          return item;
      }
    }),
    wrapperClass: 'grid-cols-1 md:grid-cols-2',
  };
});

const deviceFormOptions = computed<VbenFormProps>(() => {
  const base = buildDeviceFormOptions($t);
  return {
    ...base,
    schema: (base.schema || []).map((item: any) => {
      switch (item.fieldName) {
        case 'modelId':
          return {
            ...item,
            component: 'Select',
            componentProps: {
              ...(resolveComponentProps(item) || {}),
              allowClear: true,
              loading: modelLoading.value,
              options: deviceModelDialogOptions.value,
              placeholder: $t('page.device.selectModelId'),
              showSearch: true,
            },
          };
        case 'useStatus':
          return {
            ...item,
            component: 'Select',
            componentProps: {
              ...(resolveComponentProps(item) || {}),
              allowClear: true,
              options: deviceUseStatusOptions.value,
              placeholder: $t('page.device.selectUseStatus'),
            },
          };
        default:
          return item;
      }
    }),
    wrapperClass: 'grid-cols-1 md:grid-cols-2',
  };
});

const groupFormFieldNames = computed(() =>
  (groupFormOptions.value.schema || []).map((item: any) => item.fieldName),
);
const deviceFormFieldNames = computed(() =>
  (deviceFormOptions.value.schema || []).map((item: any) => item.fieldName),
);

const groupTreeData = computed<TreeProps['treeData']>(() =>
  groupTreeItems.value.map((item) => toGroupTreeNode(item)),
);

const [DeviceGrid, deviceGridApi] = useVbenVxeGrid<DeviceRow>({
  gridClass: 'xdev-device-center-grid',
  gridOptions: deviceGridOptions,
});

function resolveComponentProps(item: any) {
  if (typeof item.componentProps === 'function') {
    return item.componentProps({});
  }
  return item.componentProps || {};
}

function resetGroupFormModel() {
  for (const key of groupFormFieldNames.value) {
    groupFormModel[key] = undefined;
  }
  groupFormModel.parentId = selectedGroupId.value;
  groupFormModel.sortOrder = 0;
  groupFormModel.status = 'ON';
  groupFormModel.visible = false;
}

function resetDeviceFormModel() {
  for (const key of deviceFormFieldNames.value) {
    deviceFormModel[key] = undefined;
  }
  deviceFormModel.useStatus = 'USING';
  deviceFormModel.modelId = undefined;
}

function resetDeviceSearchForm() {
  deviceSearchForm.deviceCode = '';
  deviceSearchForm.modelId = undefined;
  deviceSearchForm.name = '';
  deviceSearchForm.useStatus = undefined;
}

function resetBindSearchForm() {
  bindSearchForm.deviceCode = '';
  bindSearchForm.modelId = undefined;
  bindSearchForm.name = '';
  bindSearchForm.useStatus = undefined;
  bindPagination.current = 1;
}

function toGroupTreeNode(group: AdminDeviceGroup): NonNullable<TreeProps['treeData']>[number] {
  return {
    key: String(group.id ?? group.groupName ?? Math.random()),
    title: getGroupTreeNodeTitle(group),
    children: (group.children ?? []).map((child) => toGroupTreeNode(child as AdminDeviceGroup)),
  };
}

function buildParentTreeOptions(
  items: AdminDeviceGroup[],
  excludeId?: number,
): TreeOption[] {
  return items.flatMap((item) => {
    if (item.id === undefined || item.id === excludeId) {
      return [];
    }
    const children = buildParentTreeOptions(
      (item.children ?? []) as AdminDeviceGroup[],
      excludeId,
    );
    return [
      {
        children: children.length > 0 ? children : undefined,
        key: item.id,
        title: item.groupName ?? `#${item.id}`,
        value: item.id,
      },
    ];
  });
}

function findGroupById(
  items: AdminDeviceGroup[],
  id?: number,
): AdminDeviceGroup | undefined {
  if (!id) {
    return undefined;
  }
  for (const item of items) {
    if (item.id === id) {
      return item;
    }
    const child = findGroupById((item.children ?? []) as AdminDeviceGroup[], id);
    if (child) {
      return child;
    }
  }
  return undefined;
}

function buildOrgUnitTreeOptions(items: AdminOrgUnitOption[]): DataNode[] {
  const nodeMap = new Map<number, DataNode>();
  const parentMap = new Map<number, number | undefined>();
  const roots: DataNode[] = [];

  for (const item of items) {
    if (typeof item.raw.id !== 'number') {
      continue;
    }
    nodeMap.set(item.raw.id, {
      key: item.raw.id,
      title: item.label,
      value: item.raw.id,
    });
    parentMap.set(item.raw.id, item.raw.parentId);
  }

  for (const [id, node] of nodeMap) {
    const parentID = parentMap.get(id);
    const parent = parentID ? nodeMap.get(parentID) : undefined;
    if (parent) {
      const children = (parent.children ?? []) as DataNode[];
      children.push(node);
      parent.children = children;
    } else {
      roots.push(node);
    }
  }

  return roots;
}

function resolveDeviceModelLabel(value: number | string | undefined) {
  if (value === undefined || value === null || value === '') {
    return '-';
  }
  return deviceModelOptionsMap.value.get(value) ?? value;
}

function resolveUseStatusText(value: string | undefined) {
  if (!value) {
    return '-';
  }
  return deviceUseStatusTextMap.value[value as keyof typeof deviceUseStatusTextMap.value] ?? value;
}

function resolveGroupTypeText(value: string | undefined) {
  if (!value) {
    return '-';
  }
  return deviceTypeTextMap.value[value as keyof typeof deviceTypeTextMap.value] ?? value;
}

function formatTime(value?: string) {
  return value ? dayjs(value).format('YYYY-MM-DD HH:mm') : '-';
}

function resolveOrgUnitDisplayName(item: AdminOrgUnitOption['raw']) {
  return item.name ?? item.code ?? (item.id ? `#${item.id}` : '-');
}

function resolveUserDisplayName(item: AdminUserOption['raw']) {
  return item.realname || item.nickname || item.username || (item.id ? `#${item.id}` : '-');
}

function handleOrgUnitTreeCheck(
  checked: Key[] | { checked: Key[]; halfChecked: Key[] },
) {
  const rawKeys = Array.isArray(checked) ? checked : checked.checked;
  selectedOrgUnitIDs.value = rawKeys
    .map((item) => Number(item))
    .filter((item) => Number.isFinite(item));
}

function getBoundDeviceIDs(items: AdminDeviceGroupDevice[]) {
  return items
    .map((item) => item.deviceId)
    .filter((id): id is number => typeof id === 'number');
}

function toBackendSortField(sortField: string) {
  return sortField.replaceAll(/[A-Z]/g, (match) => `_${match.toLowerCase()}`);
}

function canMutateTenant(tenantId?: number | null) {
  return (tenantId ?? 0) === currentTenantId.value;
}

function canMutateDevice(record: DeviceRow) {
  return canMutateTenant(record.tenantId);
}

function normalizeBindSortField(
  field?: string | number | ReadonlyArray<string | number>,
) {
  if (Array.isArray(field)) {
    return normalizeAdminTableSortField([...field]);
  }
  if (field === undefined) {
    return normalizeAdminTableSortField(undefined);
  }
  return normalizeAdminTableSortField(field as string | number);
}

async function loadGroups() {
  deviceGroupLoading.value = true;
  try {
    const response = await listDeviceGroupPage({
      page: 1,
      pageSize: 500,
    });
    groupTreeItems.value = response.items;

    const currentExists = findGroupById(response.items, selectedGroupId.value);
    if (!currentExists) {
      selectedGroupId.value = findFirstGroupId(response.items);
    }
    expandedGroupKeys.value = getGroupTreeKeys(response.items);
  } catch (error) {
    message.error(
      (error as Error).message || $t('page.deviceCenter.loadGroupsFailed'),
    );
  } finally {
    deviceGroupLoading.value = false;
  }
}

async function loadDeviceModels(tenantId?: number) {
  modelLoading.value = true;
  try {
    deviceModelDialogOptions.value = await listDeviceModelOptions(tenantId);
  } catch (error) {
    message.error(
      (error as Error).message || $t('page.deviceCenter.loadDeviceModelsFailed'),
    );
  } finally {
    modelLoading.value = false;
  }
}

async function loadStaticDeviceModels() {
  try {
    deviceModelOptions.value = await listDeviceModelOptions();
  } catch (error) {
    message.error(
      (error as Error).message || $t('page.deviceCenter.loadDeviceModelsFailed'),
    );
  }
}

async function loadOrgUnitOptions(tenantId?: number) {
  const response = await listAdminOrgUnitsApi({
    page: 1,
    pageSize: 500,
    sorting: [{ direction: 'ASC', field: 'sort_order' }],
    tenantId,
  });
  orgUnitDialogOptions.value = (response.items ?? []).map((item) => ({
    label: resolveOrgUnitDisplayName(item),
    raw: item,
    value: item.id ?? 0,
  }));
}

async function loadUserOptions(tenantId?: number, orgUnitId?: number) {
  const response = await listAdminUsersApi({
    orgUnitId,
    page: 1,
    pageSize: 500,
    sorting: [{ direction: 'ASC', field: 'username' }],
    tenantId,
  });
  userDialogOptions.value = (response.items ?? []).map((item) => ({
    label: resolveUserDisplayName(item),
    raw: item,
    value: item.id ?? 0,
  }));
}

async function loadCurrentGroupRelations() {
  if (!selectedGroupId.value) {
    currentGroupRelationItems.value = [];
    boundDeviceIdList.value = [];
    return;
  }
  const relationResponse = await listDeviceGroupDevicePage({
    groupId: String(selectedGroupId.value),
    page: 1,
    pageSize: 1000,
  });
  currentGroupRelationItems.value = relationResponse.items ?? [];
  boundDeviceIdList.value = getBoundDeviceIDs(currentGroupRelationItems.value);
}

async function loadCurrentGroupOrgUnits() {
  if (!selectedGroupId.value) {
    currentGroupOrgUnitItems.value = [];
    groupOrgUnitRows.value = [];
    return;
  }
  const response = await listDeviceGroupOrgUnitPage({
    groupId: String(selectedGroupId.value),
    page: 1,
    pageSize: 500,
  });
  currentGroupOrgUnitItems.value = response.items ?? [];

  const displayResponse = await listAdminOrgUnitsApi({
    page: 1,
    pageSize: 500,
    sorting: [{ direction: 'ASC', field: 'sort_order' }],
    tenantId: selectedGroup.value?.tenantId ?? undefined,
  });
  const optionMap = new Map(
    (displayResponse.items ?? []).map((item) => [item.id ?? 0, item]),
  );
  groupOrgUnitRows.value = currentGroupOrgUnitItems.value
    .map((item) => {
      const option = item.orgUnitId ? optionMap.get(item.orgUnitId) : undefined;
      return {
        id: item.orgUnitId ?? 0,
        name: option ? resolveOrgUnitDisplayName(option) : `#${item.orgUnitId ?? 0}`,
        relationId: item.id,
      };
    })
    .filter((item) => item.id > 0);
}

async function loadCurrentGroupUsers() {
  if (!selectedGroupId.value) {
    currentGroupUserItems.value = [];
    groupUserRows.value = [];
    return;
  }
  const response = await listDeviceGroupUserPage({
    groupId: String(selectedGroupId.value),
    page: 1,
    pageSize: 500,
  });
  currentGroupUserItems.value = response.items ?? [];

  const displayResponse = await listAdminUsersApi({
    page: 1,
    pageSize: 500,
    sorting: [{ direction: 'ASC', field: 'username' }],
    tenantId: selectedGroup.value?.tenantId ?? undefined,
  });
  const optionMap = new Map(
    (displayResponse.items ?? []).map((item) => [item.id ?? 0, item]),
  );
  groupUserRows.value = currentGroupUserItems.value
    .map((item) => {
      const option = item.userId ? optionMap.get(item.userId) : undefined;
      return {
        id: item.userId ?? 0,
        label: option ? resolveUserDisplayName(option) : `#${item.userId ?? 0}`,
        orgUnitNames: option?.orgUnitNames ?? [],
        relationId: item.id,
      };
    })
    .filter((item) => item.id > 0);
}

async function deleteCurrentGroupRelationsForGroup() {
  if (!selectedGroupId.value) {
    return;
  }
  await loadCurrentGroupRelations();
  await loadCurrentGroupOrgUnits();
  await loadCurrentGroupUsers();

  for (const relation of currentGroupRelationItems.value) {
    if (relation?.id) {
      await deleteDeviceGroupDevice(relation.id);
    }
  }
  for (const relation of currentGroupOrgUnitItems.value) {
    if (relation?.id) {
      await deleteDeviceGroupOrgUnit(relation.id);
    }
  }
  for (const relation of currentGroupUserItems.value) {
    if (relation?.id) {
      await deleteDeviceGroupUser(relation.id);
    }
  }
}

async function loadDevicePage(
  params: Pick<AdminDeviceListParams, 'page' | 'pageSize' | 'sorting'>,
): Promise<AdminDeviceListResult> {
  try {
    if (!selectedGroupId.value) {
      currentGroupRelationItems.value = [];
      boundDeviceIdList.value = [];
      return {
        items: [],
        total: 0,
      };
    }

    await loadCurrentGroupRelations();

    if (boundDeviceIdList.value.length === 0) {
      return {
        items: [],
        total: 0,
      };
    }

    const deviceResponse = await listDevicePage({
      ...deviceSearchForm,
      page: params.page,
      pageSize: 1000,
      sorting: params.sorting,
    });

    const filteredItems = (deviceResponse.items ?? []).filter((item) =>
      typeof item.id === 'number' ? boundDeviceIdList.value.includes(item.id) : false,
    );
    const page = params.page ?? 1;
    const pageSize = params.pageSize ?? 20;
    const start = (page - 1) * pageSize;
    const pagedItems = filteredItems.slice(start, start + pageSize);

    const relationMap = new Map<number, number | undefined>();
    for (const item of currentGroupRelationItems.value) {
      if (typeof item.deviceId === 'number') {
        relationMap.set(item.deviceId, item.id);
      }
    }

    const items = pagedItems
      .map((item) => ({
        ...item,
        relationId: item.id ? relationMap.get(item.id) : undefined,
      }));
    return {
      items,
      total: filteredItems.length,
    };
  } catch (error) {
    message.error(
      (error as Error).message || $t('page.deviceCenter.loadDevicesFailed'),
    );
    return {
      items: [],
      total: 0,
    };
  }
}

async function loadDevices() {
  await deviceGridApi.reload();
}

async function loadBindableDevices() {
  try {
    bindLoading.value = true;
    const groupTenantId = selectedGroup.value?.tenantId;
    const response = await listDevicePage({
      ...bindSearchForm,
      tenantId: groupTenantId,
      page: bindPagination.current,
      pageSize: bindPagination.pageSize,
      sorting: bindSorting.value,
    });

    const boundSet = new Set(boundDeviceIdList.value);
    bindTableRows.value = (response.items ?? []).map((item) => ({
      ...item,
      alreadyBound: typeof item.id === 'number' ? boundSet.has(item.id) : false,
    }));
    bindPagination.total = response.total ?? 0;
  } catch (error) {
    message.error(
      (error as Error).message || $t('page.deviceCenter.loadDevicesFailed'),
    );
  } finally {
    bindLoading.value = false;
  }
}

async function refreshAll() {
  await loadGroups();
  await loadStaticDeviceModels();
  await loadOrgUnitOptions();
  await loadUserOptions();
  await loadCurrentGroupOrgUnits();
  await loadCurrentGroupUsers();
  await loadDevices();
}

async function handleGroupSelect(keys: (number | string)[]) {
  const key = keys[0];
  selectedGroupId.value = key ? Number.parseInt(String(key), 10) : undefined;
  await loadCurrentGroupOrgUnits();
  await loadCurrentGroupUsers();
  await loadDevices();
}

async function handleSearchDevices() {
  await loadDevices();
}

async function handleResetDeviceSearch() {
  resetDeviceSearchForm();
  await loadDevices();
}

async function openCreateGroup() {
  editingGroupId.value = undefined;
  resetGroupFormModel();
  groupModalOpen.value = true;
  await nextTick();
  groupFormRef.value?.clearValidate();
}

async function openEditGroup() {
  if (!selectedGroupId.value) {
    message.warning($t('page.deviceCenter.selectGroupRequired'));
    return;
  }

  const group = await getDeviceGroupById(selectedGroupId.value);
  if (!group) {
    message.warning($t('page.deviceCenter.groupNotFound'));
    return;
  }

  editingGroupId.value = group.id;
  Object.assign(groupFormModel, group);
  groupModalOpen.value = true;
  await nextTick();
  groupFormRef.value?.clearValidate();
}

async function submitGroup() {
  await groupFormRef.value?.validate();
  groupSubmitting.value = true;
  try {
    if (editingGroupId.value) {
      await updateDeviceGroup(editingGroupId.value, groupFormModel);
      message.success($t('page.deviceGroup.updateSuccess'));
    } else {
      await createDeviceGroup(groupFormModel);
      message.success($t('page.deviceGroup.createSuccess'));
    }
    groupModalOpen.value = false;
    await loadGroups();
    await loadDevices();
  } finally {
    groupSubmitting.value = false;
  }
}

async function handleDeleteGroup() {
  if (!selectedGroupId.value) {
    message.warning($t('page.deviceCenter.selectGroupRequired'));
    return;
  }
  await deleteCurrentGroupRelationsForGroup();
  await deleteDeviceGroup(selectedGroupId.value);
  message.success($t('page.deviceGroup.deleteSuccess'));
  selectedGroupId.value = undefined;
  await refreshAll();
}

async function openCreateDevice() {
  if (!selectedGroupId.value) {
    message.warning($t('page.deviceCenter.selectGroupRequired'));
    return;
  }
  if (!selectedGroupIsLeaf.value) {
    message.warning($t('page.deviceCenter.selectGroupRequired'));
    return;
  }
  if (!selectedGroupMutable.value) {
    message.warning($t('common.detail'));
    return;
  }
  editingDeviceId.value = undefined;
  editingDeviceReadonly.value = false;
  resetDeviceFormModel();
  await loadDeviceModels(selectedGroup.value?.tenantId ?? undefined);
  deviceModalOpen.value = true;
  await nextTick();
  deviceFormRef.value?.clearValidate();
}

async function openEditDevice(record: DeviceRow) {
  if (!record.id) {
    message.warning($t('common.detail'));
    return;
  }
  try {
    const data = await getDeviceById(record.id);
    if (!data) {
      message.warning($t('page.deviceCenter.deviceNotFound'));
      return;
    }
    editingDeviceId.value = record.id;
    editingDeviceReadonly.value = !canMutateDevice(record);
    await loadDeviceModels(data.tenantId ?? undefined);
    resetDeviceFormModel();
    Object.assign(deviceFormModel, data);
    deviceModalOpen.value = true;
    await nextTick();
    deviceFormRef.value?.clearValidate();
  } catch (error) {
    message.error((error as Error)?.message || $t('common.detail'));
  }
}

async function ensureDeviceLinkedToCurrentGroup(deviceCode?: string) {
  if (!selectedGroupId.value || !deviceCode?.trim()) {
    return;
  }

  const deviceResponse = await listDevicePage({
    deviceCode: deviceCode.trim(),
    page: 1,
    pageSize: 20,
  });
  const createdDevice = (deviceResponse.items ?? []).find(
    (item) => item.deviceCode?.trim() === deviceCode.trim(),
  );
  if (!createdDevice?.id) {
    return;
  }

  const existing = await listDeviceGroupDevicePage({
    deviceId: String(createdDevice.id),
    groupId: String(selectedGroupId.value),
    page: 1,
    pageSize: 20,
  });
  if ((existing.items ?? []).length > 0) {
    return;
  }

  await createDeviceGroupDevice({
    deviceId: createdDevice.id,
    groupId: selectedGroupId.value,
  } as AdminDeviceGroupDeviceSaveInput);
}

async function submitDevice() {
  if (editingDeviceReadonly.value) {
    deviceModalOpen.value = false;
    return;
  }
  await deviceFormRef.value?.validate();
  deviceSubmitting.value = true;
  try {
    if (editingDeviceId.value) {
      await updateDevice(editingDeviceId.value, deviceFormModel);
      message.success($t('page.device.updateSuccess'));
    } else {
      await createDevice(deviceFormModel);
      await ensureDeviceLinkedToCurrentGroup(deviceFormModel.deviceCode);
      message.success($t('page.device.createSuccess'));
    }
    deviceModalOpen.value = false;
    await loadDevices();
  } finally {
    deviceSubmitting.value = false;
  }
}

async function handleDeleteDevice(record: DeviceRow) {
  if (!record.id || !selectedGroupId.value || !canMutateDevice(record)) {
    return;
  }
  const relationResponse = await listDeviceGroupDevicePage({
    deviceId: String(record.id),
    page: 1,
    pageSize: 1000,
  });

  for (const relation of relationResponse.items ?? []) {
    if (relation?.id) {
      await deleteDeviceGroupDevice(relation.id);
    }
  }

  await deleteDevice(record.id);
  message.success($t('page.deviceCenter.deleteDeviceSuccess'));
  await loadDevices();
}

async function handleUnbindDevice(record: DeviceRow) {
  if (!record.relationId || !canMutateDevice(record)) {
    return;
  }
  await deleteDeviceGroupDevice(record.relationId);
  message.success($t('authentication.unbindAction'));
  await loadDevices();
}

async function openBindModal() {
  if (!selectedGroupId.value) {
    message.warning($t('page.deviceCenter.selectGroupRequired'));
    return;
  }
  if (!selectedGroupMutable.value) {
    message.warning($t('common.detail'));
    return;
  }
  bindModalOpen.value = true;
  bindSelectedRowKeys.value = [];
  resetBindSearchForm();
  await loadCurrentGroupRelations();
  await loadBindableDevices();
}

async function handleBindSearch() {
  bindPagination.current = 1;
  await loadBindableDevices();
}

async function handleBindReset() {
  resetBindSearchForm();
  bindSelectedRowKeys.value = [];
  bindSorting.value = undefined;
  await loadBindableDevices();
}

async function handleBindTableChange(
  pagination: TablePaginationConfig,
  _: Record<string, unknown>,
  sorter:
    | {
        field?: string | number | ReadonlyArray<string | number>;
        order?: 'ascend' | 'descend' | null;
      }
    | Array<{
        field?: string | number | ReadonlyArray<string | number>;
        order?: 'ascend' | 'descend' | null;
      }>,
): Promise<void> {
  bindPagination.current = pagination.current ?? 1;
  bindPagination.pageSize = pagination.pageSize ?? 10;
  const firstSorter = Array.isArray(sorter) ? sorter[0] : sorter;
  const direction = normalizeAdminTableSortDirection(firstSorter?.order);
  const field = normalizeBindSortField(firstSorter?.field);
  bindSorting.value =
    direction && field
      ? [
          {
            direction,
            field,
          },
        ]
      : undefined;
  await loadBindableDevices();
}

async function submitBinding() {
  if (!selectedGroupId.value) {
    message.warning($t('page.deviceCenter.selectGroupRequired'));
    return;
  }
  if (bindSelectedRowKeys.value.length === 0) {
    message.warning($t('page.deviceGroupDevice.selectFilterDeviceId'));
    return;
  }

  bindingSubmitting.value = true;
  try {
    const groupTenantId = selectedGroup.value?.tenantId;
    const selectedRows = bindTableRows.value.filter(
      (item) =>
        item.id !== undefined && bindSelectedRowKeys.value.includes(item.id),
    );
    const invalidRow = selectedRows.find(
      (item) => (item.tenantId ?? 0) !== (groupTenantId ?? 0),
    );
    if (invalidRow) {
      message.warning($t('common.detail'));
      return;
    }
    for (const deviceId of bindSelectedRowKeys.value) {
      await createDeviceGroupDevice({
        deviceId,
        groupId: selectedGroupId.value,
      } as AdminDeviceGroupDeviceSaveInput);
    }
    message.success($t('page.deviceGroupDevice.createSuccess'));
    bindModalOpen.value = false;
    bindSelectedRowKeys.value = [];
    await loadDevices();
  } finally {
    bindingSubmitting.value = false;
  }
}

async function openGroupOrgUnitModal() {
  if (!selectedGroupId.value || !selectedGroupMutable.value) {
    message.warning($t('page.deviceCenter.selectGroupRequired'));
    return;
  }
  const tenantId = selectedGroup.value?.tenantId ?? undefined;
  await loadOrgUnitOptions(tenantId);
  await loadCurrentGroupOrgUnits();
  selectedOrgUnitIDs.value = currentGroupOrgUnitItems.value
    .map((item) => item.orgUnitId)
    .filter((value): value is number => typeof value === 'number');
  groupOrgUnitModalOpen.value = true;
}

async function submitGroupOrgUnitBinding() {
  if (!selectedGroupId.value) {
    return;
  }
  groupOrgUnitSubmitting.value = true;
  try {
    const existingMap = new Map<number, number>();
    for (const item of currentGroupOrgUnitItems.value) {
      if (typeof item.orgUnitId === 'number' && typeof item.id === 'number') {
        existingMap.set(item.orgUnitId, item.id);
      }
    }
    const nextSet = new Set(selectedOrgUnitIDs.value);

    for (const item of currentGroupOrgUnitItems.value) {
      if (
        typeof item.orgUnitId === 'number' &&
        typeof item.id === 'number' &&
        !nextSet.has(item.orgUnitId)
      ) {
        await deleteDeviceGroupOrgUnit(item.id);
      }
    }

    for (const orgUnitId of nextSet) {
      if (existingMap.has(orgUnitId)) {
        continue;
      }
      await createDeviceGroupOrgUnit({
        groupId: selectedGroupId.value,
        orgUnitId,
      });
    }

    groupOrgUnitModalOpen.value = false;
    await loadCurrentGroupOrgUnits();
    message.success($t('page.deviceGroupOrgUnit.updateSuccess') || $t('common.saveSuccess'));
  } finally {
    groupOrgUnitSubmitting.value = false;
  }
}

async function openGroupUserModal() {
  if (!selectedGroupId.value || !selectedGroupMutable.value) {
    message.warning($t('page.deviceCenter.selectGroupRequired'));
    return;
  }
  const tenantId = selectedGroup.value?.tenantId ?? undefined;
  userFilterOrgUnitId.value = undefined;
  await loadOrgUnitOptions(tenantId);
  await loadUserOptions(tenantId);
  await loadCurrentGroupUsers();
  selectedUserIDs.value = currentGroupUserItems.value
    .map((item) => item.userId)
    .filter((value): value is number => typeof value === 'number');
  groupUserModalOpen.value = true;
}

async function handleUserFilterOrgUnitChange(value: unknown) {
  const rawValue = Array.isArray(value) ? value[0] : value;
  const nextValue =
    rawValue === undefined || rawValue === null || rawValue === ''
      ? undefined
      : Number(rawValue);
  userFilterOrgUnitId.value = nextValue;
  await loadUserOptions(selectedGroup.value?.tenantId ?? undefined, nextValue);
}

async function submitGroupUserBinding() {
  if (!selectedGroupId.value) {
    return;
  }
  groupUserSubmitting.value = true;
  try {
    const existingMap = new Map<number, number>();
    for (const item of currentGroupUserItems.value) {
      if (typeof item.userId === 'number' && typeof item.id === 'number') {
        existingMap.set(item.userId, item.id);
      }
    }
    const nextSet = new Set(selectedUserIDs.value);

    for (const item of currentGroupUserItems.value) {
      if (
        typeof item.userId === 'number' &&
        typeof item.id === 'number' &&
        !nextSet.has(item.userId)
      ) {
        await deleteDeviceGroupUser(item.id);
      }
    }

    for (const userId of nextSet) {
      if (existingMap.has(userId)) {
        continue;
      }
      await createDeviceGroupUser({
        groupId: selectedGroupId.value,
        userId,
      });
    }

    groupUserModalOpen.value = false;
    await loadCurrentGroupUsers();
    message.success($t('page.deviceGroupUser.updateSuccess') || $t('common.saveSuccess'));
  } finally {
    groupUserSubmitting.value = false;
  }
}

function getGroupTreeNodeTitle(group: AdminDeviceGroup) {
  const label = group.groupName ?? `#${group.id}`;
  const typeText = resolveGroupTypeText(group.type);
  return typeText !== '-' ? `${label} 路 ${typeText}` : label;
}

function getGroupTreeKeys(items: AdminDeviceGroup[]): Array<number | string> {
  const keys: Array<number | string> = [];
  const visit = (nodes: AdminDeviceGroup[]) => {
    for (const node of nodes) {
      if (node.id !== undefined) {
        keys.push(node.id);
      }
      const children = (node.children ?? []) as AdminDeviceGroup[];
      if (children.length > 0) {
        visit(children);
      }
    }
  };
  visit(items);
  return keys;
}

function toggleGroupTreeExpand() {
  const allKeys = selectedGroupAllKeys.value;
  expandedGroupKeys.value =
    expandedGroupKeys.value.length === allKeys.length ? [] : allKeys;
}

function handleExpandedKeysChange(keys: (number | string)[]) {
  expandedGroupKeys.value = keys;
}

function findFirstGroupId(items: AdminDeviceGroup[]): number | undefined {
  for (const item of items) {
    if (item.id !== undefined) {
      return item.id;
    }
    const childId = findFirstGroupId((item.children ?? []) as AdminDeviceGroup[]);
    if (childId !== undefined) {
      return childId;
    }
  }
  return undefined;
}

onMounted(() => {
  void refreshAll();
  void loadDeviceModels();
});
</script>

<template>
  <Page auto-content-height :title="$t('menu.xdev.deviceCenter')">
    <div class="xdev-device-center">
      <aside class="xdev-device-center__group">
        <div class="panel-header">
          <div>
            <div class="panel-title">{{ $t('page.deviceGroup.moduleName') }}</div>
          </div>
          <Space>
            <Button size="small" type="text" @click="toggleGroupTreeExpand">
              <template #icon>
                <IconifyIcon
                  :icon="
                    expandedGroupKeys.length === selectedGroupAllKeys.length
                      ? 'lucide:fold-vertical'
                      : 'lucide:unfold-vertical'
                  "
                />
              </template>
            </Button>
            <Button size="small" type="text" @click="openCreateGroup">
              <template #icon>
                <IconifyIcon icon="lucide:plus" />
              </template>
            </Button>
            <Button :disabled="!selectedGroupId" size="small" type="text" @click="openEditGroup">
              <template #icon>
                <IconifyIcon icon="lucide:pencil" />
              </template>
            </Button>
            <Popconfirm
              :title="$t('ui.actionMessage.deleteConfirm', [$t('page.deviceGroup.moduleName')])"
              @confirm="handleDeleteGroup"
            >
              <Button :disabled="!selectedGroupId" danger size="small" type="text">
                <template #icon>
                  <IconifyIcon icon="lucide:trash-2" />
                </template>
              </Button>
            </Popconfirm>
          </Space>
        </div>

        <Tree
          :expanded-keys="expandedGroupKeys"
          :loading="deviceGroupLoading"
          :selected-keys="selectedGroupTreeKeys"
          :tree-data="groupTreeData"
          @update:expanded-keys="handleExpandedKeysChange"
          @select="handleGroupSelect"
        >
          <template #title="{ title }">
            <span class="group-tree__title">{{ title }}</span>
          </template>
        </Tree>
      </aside>

      <section class="xdev-device-center__device">
        <div class="device-search-card">
          <div class="device-banner">
            <div class="device-banner__label">{{ $t('page.deviceGroup.moduleName') }}</div>
            <div class="device-banner__value">{{ selectedGroupLabel }}</div>
            <div class="device-banner__meta">{{ selectedGroupType }}</div>
            <div class="device-banner__meta">
              {{ $t('page.device.moduleName') }} {{ selectedGroupBoundCount }}
            </div>
            <div class="device-banner__relation">
              <span class="device-banner__relation-label">
                {{ $t('page.orgUnit.orgUnit') }}:
              </span>
              <div class="device-banner__relation-items">
                <template v-if="groupOrgUnitRows.length > 0">
                  <span
                    v-for="item in groupOrgUnitRows.slice(0, 3)"
                    :key="item.id"
                    class="device-banner__relation-chip"
                  >
                    <IconifyIcon class="device-banner__relation-icon" icon="lucide:building-2" />
                    <span>{{ item.name }}</span>
                  </span>
                  <span v-if="groupOrgUnitRows.length > 3" class="device-banner__relation-more">
                    +{{ groupOrgUnitRows.length - 3 }}
                  </span>
                </template>
                <span v-else class="device-banner__relation-empty">-</span>
                <Button
                  v-if="selectedGroupIsLeaf"
                  class="device-banner__relation-action"
                  :disabled="!selectedGroupId || !selectedGroupMutable"
                  size="small"
                  type="link"
                  @click="openGroupOrgUnitModal"
                >
                  {{ $t('common.edit') }}
                </Button>
              </div>
            </div>
            <div class="device-banner__relation">
              <span class="device-banner__relation-label">
                {{ $t('page.user.identity') }}:
              </span>
              <div class="device-banner__relation-items">
                <template v-if="groupUserRows.length > 0">
                  <span
                    v-for="item in groupUserRows.slice(0, 3)"
                    :key="item.id"
                    class="device-banner__relation-chip"
                  >
                    <IconifyIcon class="device-banner__relation-icon" icon="lucide:user-round" />
                    <span>{{ item.label }}</span>
                  </span>
                  <span v-if="groupUserRows.length > 3" class="device-banner__relation-more">
                    +{{ groupUserRows.length - 3 }}
                  </span>
                </template>
                <span v-else class="device-banner__relation-empty">-</span>
                <Button
                  v-if="selectedGroupIsLeaf"
                  class="device-banner__relation-action"
                  :disabled="!selectedGroupId || !selectedGroupMutable"
                  size="small"
                  type="link"
                  @click="openGroupUserModal"
                >
                  {{ $t('common.edit') }}
                </Button>
              </div>
            </div>
          </div>

          <div class="device-search-form">
            <Input
              v-model:value="deviceSearchForm.deviceCode"
              allow-clear
              :placeholder="$t('page.device.searchFilterDeviceCode')"
            />
            <Input
              v-model:value="deviceSearchForm.name"
              allow-clear
              :placeholder="$t('page.device.searchFilterName')"
            />
            <Select
              v-model:value="deviceSearchForm.modelId"
              allow-clear
              :loading="modelLoading"
              :options="deviceModelOptions"
              :placeholder="$t('page.device.selectFilterModelId')"
              show-search
            />
            <Select
              v-model:value="deviceSearchForm.useStatus"
              allow-clear
              :options="deviceUseStatusOptions"
              :placeholder="$t('page.device.selectFilterUseStatus')"
            />
            <Space>
              <Button @click="handleResetDeviceSearch">
                {{ $t('common.reset') }}
              </Button>
              <Button type="primary" @click="handleSearchDevices">
                {{ $t('common.search') }}
              </Button>
            </Space>
          </div>
        </div>

        <DeviceGrid :table-title="$t('page.device.moduleName')">
          <template #toolPrefix>
            <Space>
              <Button
                :disabled="!selectedGroupId || !selectedGroupIsLeaf || !selectedGroupMutable"
                type="primary"
                @click="openCreateDevice"
              >
                <template #icon>
                  <IconifyIcon icon="lucide:plus" />
                </template>
                {{ $t('common.create') }}
              </Button>
              <Button
                :disabled="!selectedGroupId || !selectedGroupIsLeaf || !selectedGroupMutable"
                @click="openBindModal"
              >
                <template #icon>
                  <IconifyIcon icon="lucide:link" />
                </template>
                {{ $t('page.deviceGroup.binding') }}
              </Button>
            </Space>
          </template>

          <template #modelId="{ row }">
            {{ resolveDeviceModelLabel(row.modelId) }}
          </template>

          <template #useStatus="{ row }">
            {{ resolveUseStatusText(row.useStatus) }}
          </template>

          <template #createdAt="{ row }">
            {{ formatTime(row.createdAt as string | undefined) }}
          </template>

          <template #action="{ row }">
            <Space>
              <Button size="small" type="link" @click="openEditDevice(row)">
                {{ canMutateDevice(row) ? $t('common.edit') : $t('common.detail') }}
              </Button>
              <Popconfirm
                :title="$t('ui.actionMessage.deleteConfirm', [$t('authentication.unbindAction')])"
                @confirm="handleUnbindDevice(row)"
              >
                <Button :disabled="!row.relationId || !canMutateDevice(row)" size="small" type="link">
                  {{ $t('authentication.unbindAction') }}
                </Button>
              </Popconfirm>
              <Popconfirm
                :title="$t('ui.actionMessage.deleteConfirm', [$t('page.device.moduleName')])"
                @confirm="handleDeleteDevice(row)"
              >
                <Button :disabled="!canMutateDevice(row)" danger size="small" type="link">
                  {{ $t('common.delete') }}
                </Button>
              </Popconfirm>
            </Space>
          </template>
        </DeviceGrid>
      </section>
    </div>

    <Modal
      v-model:open="groupModalOpen"
      destroy-on-close
      :confirm-loading="groupSubmitting"
      :title="editingGroupId ? $t('page.deviceGroup.editTitle') : $t('page.deviceGroup.createTitle')"
      width="720px"
      @ok="submitGroup"
    >
      <Form ref="groupFormRef" :model="groupFormModel" :label-col="{ span: 5 }">
        <AdminGeneratedForm :model="groupFormModel" :schema="groupFormOptions.schema || []" />
      </Form>
    </Modal>

    <Modal
      v-model:open="deviceModalOpen"
      destroy-on-close
      :confirm-loading="deviceSubmitting"
      :ok-button-props="{ disabled: editingDeviceReadonly }"
      :title="deviceModalTitle"
      width="720px"
      @ok="submitDevice"
    >
      <Form ref="deviceFormRef" :model="deviceFormModel" :label-col="{ span: 5 }">
        <AdminGeneratedForm :model="deviceFormModel" :schema="deviceFormOptions.schema || []" />
      </Form>
    </Modal>

    <Modal
      v-model:open="bindModalOpen"
      destroy-on-close
      :confirm-loading="bindingSubmitting"
      :title="$t('page.deviceGroup.binding')"
      width="1100px"
      @ok="submitBinding"
    >
      <div class="bind-search-form">
        <Input
          v-model:value="bindSearchForm.deviceCode"
          allow-clear
          :placeholder="$t('page.device.searchFilterDeviceCode')"
        />
        <Input
          v-model:value="bindSearchForm.name"
          allow-clear
          :placeholder="$t('page.device.searchFilterName')"
        />
        <Select
          v-model:value="bindSearchForm.modelId"
          allow-clear
          :loading="modelLoading"
          :options="deviceModelOptions"
          :placeholder="$t('page.device.selectFilterModelId')"
          show-search
        />
        <Select
          v-model:value="bindSearchForm.useStatus"
          allow-clear
          :options="deviceUseStatusOptions"
          :placeholder="$t('page.device.selectFilterUseStatus')"
        />
        <Space>
          <Button @click="handleBindReset">
            {{ $t('common.reset') }}
          </Button>
          <Button type="primary" @click="handleBindSearch">
            {{ $t('common.search') }}
          </Button>
        </Space>
      </div>

      <Table
        bordered
        class="bind-device-table"
        :columns="bindTableColumns"
        :data-source="bindTableRows"
        :loading="bindLoading"
        :pagination="{
          current: bindPagination.current,
          pageSize: bindPagination.pageSize,
          total: bindPagination.total,
          showSizeChanger: true,
        }"
        :row-key="(record) => record.id || 0"
        :row-selection="bindRowSelection"
        @change="handleBindTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'modelId'">
            {{ resolveDeviceModelLabel(record.modelId) }}
          </template>
          <template v-else-if="column.key === 'useStatus'">
            {{ resolveUseStatusText(record.useStatus) }}
          </template>
          <template v-else-if="column.key === 'createdAt'">
            {{ formatTime(record.createdAt as string | undefined) }}
          </template>
        </template>
      </Table>
    </Modal>

    <Modal
      v-model:open="groupOrgUnitModalOpen"
      destroy-on-close
      :confirm-loading="groupOrgUnitSubmitting"
      :title="$t('page.deviceGroupOrgUnit.moduleName')"
      width="720px"
      @ok="submitGroupOrgUnitBinding"
    >
      <Tree
        checkable
        default-expand-all
        :checked-keys="selectedOrgUnitIDs"
        :tree-data="orgUnitTreeOptions"
        @check="handleOrgUnitTreeCheck"
      />
    </Modal>

    <Modal
      v-model:open="groupUserModalOpen"
      destroy-on-close
      :confirm-loading="groupUserSubmitting"
      :title="$t('page.deviceGroupUser.moduleName')"
      width="860px"
      @ok="submitGroupUserBinding"
    >
      <div class="group-user-filter">
        <Select
          v-model:value="userFilterOrgUnitId"
          allow-clear
          :options="orgUnitDialogOptions"
          :placeholder="$t('page.user.selectFilterOrgUnitId')"
          show-search
          @change="handleUserFilterOrgUnitChange"
        />
      </div>
      <Select
        v-model:value="selectedUserIDs"
        mode="multiple"
        :options="userDialogOptions"
        :placeholder="$t('page.deviceGroupUser.selectUserId')"
        show-search
        style="width: 100%;"
      />
    </Modal>
  </Page>
</template>

<style scoped>
.xdev-device-center {
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  gap: 16px;
  min-height: 100%;
}

.xdev-device-center__group,
.xdev-device-center__device {
  min-width: 0;
  padding: 16px;
  background: hsl(var(--background));
  border: 1px solid hsl(var(--border));
  border-radius: 12px;
}

.xdev-device-center__group {
  position: sticky;
  top: 16px;
  max-height: calc(100vh - 32px);
  overflow: auto;
}

.xdev-device-center__device {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.panel-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.panel-title {
  font-size: 16px;
  font-weight: 600;
  color: hsl(var(--foreground));
}

.device-search-card {
  padding: 14px;
  background: hsl(var(--background));
  border: 1px solid hsl(var(--border));
  border-radius: 12px;
}

.device-banner {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
  padding: 12px 14px;
  margin-bottom: 14px;
  background: linear-gradient(135deg, hsl(var(--primary) / 10%), hsl(var(--accent) / 18%));
  border: 1px solid hsl(var(--primary) / 18%);
  border-radius: 10px;
}

.device-banner__label {
  font-size: 16px;
  color: hsl(var(--muted-foreground));
}

.device-banner__value {
  font-size: 16px;
  font-weight: 700;
  color: hsl(var(--foreground));
}

.device-banner__meta {
  padding: 6px 10px;
  font-size: 12px;
  color: hsl(var(--foreground));
  background: hsl(var(--background) / 72%);
  border: 1px solid hsl(var(--border));
  border-radius: 999px;
}

.device-banner__relation {
  display: flex;
  flex: 0 1 auto;
  gap: 8px;
  align-items: center;
  min-width: 0;
  padding: 6px 10px;
  background: hsl(var(--background) / 72%);
  border: 1px solid hsl(var(--border));
  border-radius: 999px;
}

.device-banner__relation-label {
  font-size: 13px;
  color: hsl(var(--muted-foreground));
}

.device-banner__relation-items {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  min-width: 0;
}

.device-banner__relation-chip {
  display: inline-flex;
  gap: 6px;
  align-items: center;
  max-width: 220px;
  padding: 2px 8px;
  overflow: hidden;
  font-size: 12px;
  color: hsl(var(--foreground));
  text-overflow: ellipsis;
  white-space: nowrap;
  background: hsl(var(--background) / 92%);
  border: 1px solid hsl(var(--border) / 80%);
  border-radius: 999px;
}

.device-banner__relation-icon {
  flex: none;
  font-size: 14px;
  color: hsl(var(--primary));
}

.device-banner__relation-more,
.device-banner__relation-empty {
  font-size: 12px;
  color: hsl(var(--muted-foreground));
}

.device-banner__relation-action {
  padding-inline: 0;
}

.device-search-form,
.bind-search-form {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 12px;
  align-items: center;
}

.group-user-filter {
  margin-bottom: 16px;
}

.group-tree__title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.bind-device-table {
  margin-top: 16px;
}

:deep(.xdev-device-center-grid) {
  min-height: 0;
}

:deep(.ant-tree .ant-tree-node-content-wrapper.ant-tree-node-selected) {
  color: hsl(var(--foreground));
  background: hsl(var(--primary) / 18%);
}

:deep(.ant-tree .ant-tree-treenode:hover .ant-tree-node-content-wrapper) {
  background: hsl(var(--accent) / 18%);
}

:deep(.ant-tree .ant-tree-node-content-wrapper.ant-tree-node-selected:hover) {
  background: hsl(var(--primary) / 24%);
}

@media (max-width: 1200px) {
  .device-search-form,
  .bind-search-form {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 900px) {
  .xdev-device-center {
    grid-template-columns: 1fr;
  }

  .xdev-device-center__group {
    position: static;
    top: auto;
    max-height: none;
  }

  .device-search-form,
  .bind-search-form {
    grid-template-columns: 1fr;
  }
}
</style>

