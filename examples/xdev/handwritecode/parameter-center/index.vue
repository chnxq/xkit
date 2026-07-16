<!-- Hand-written by Codex. -->
<script lang="ts" setup>
import type {
  FormInstance,
  MenuProps,
  TableColumnsType,
  TablePaginationConfig,
  TreeProps,
} from 'ant-design-vue';

import type { AdminDevice } from '../../provider/device.provider';
import type { AdminDeviceGroup } from '../../provider/device-group.provider';
import type { AdminDeviceModel } from '../../provider/device-model.provider';
import type { AdminDeviceModelParameterGroup } from '../../provider/device-model-parameter-group.provider';
import type { AdminDeviceModelType } from '../../provider/device-model-type.provider';
import type { AdminDeviceParameterGroup } from '../../provider/device-parameter-group.provider';
import type { AdminDeviceParameterItem } from '../../provider/device-parameter-item.provider';

import { computed, h, nextTick, onMounted, reactive, ref } from 'vue';

import { useAccess } from '@vben/access';
import { Page } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';
import { useUserStore } from '@vben/stores';

import {
  Alert,
  Button,
  Checkbox,
  Drawer,
  Dropdown,
  Empty,
  Form,
  Input,
  InputNumber,
  Menu,
  message,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tabs,
  Tag,
  Tree,
} from 'ant-design-vue';

import { $t } from '#/locales';

import { listDevicePage } from '../../provider/device.provider';
import { listDeviceGroupDevicePage } from '../../provider/device-group-device-rel.provider';
import { listDeviceGroupPage } from '../../provider/device-group.provider';
import {
  createDeviceModel,
  deleteDeviceModel,
  getDeviceModelById,
  listDeviceModelPage,
  updateDeviceModel,
} from '../../provider/device-model.provider';
import {
  createDeviceModelParameterGroup,
  deleteDeviceModelParameterGroup,
  listDeviceModelParameterGroupPage,
} from '../../provider/device-model-parameter-group.provider';
import {
  createDeviceModelType,
  deleteDeviceModelType,
  getDeviceModelTypeById,
  listDeviceModelTypePage,
  updateDeviceModelType,
} from '../../provider/device-model-type.provider';
import {
  createDeviceParameterGroup,
  deleteDeviceParameterGroup,
  getDeviceParameterGroupById,
  listDeviceParameterGroupPage,
  updateDeviceParameterGroup,
} from '../../provider/device-parameter-group.provider';
import {
  createDeviceParameterItem,
  deleteDeviceParameterItem,
  getDeviceParameterItemById,
  listDeviceParameterItemPage,
  updateDeviceParameterItem,
} from '../../provider/device-parameter-item.provider';

type ParameterGroupRow = AdminDeviceParameterGroup & {
  itemCount: number;
  modelCount: number;
  relationId?: number;
};

type AssociationGroupRow = AdminDeviceGroup & { deviceCount: number };

const PAGE_ACCESS = {
  groupCreate: ['xdev:device-parameter-group:create'],
  groupDelete: ['xdev:device-parameter-group:delete'],
  groupEdit: ['xdev:device-parameter-group:edit'],
  itemCreate: ['xdev:device-parameter-item:create'],
  itemDelete: ['xdev:device-parameter-item:delete'],
  itemEdit: ['xdev:device-parameter-item:edit'],
  modelCreate: ['xdev:device-model:create'],
  modelDelete: ['xdev:device-model:delete'],
  modelEdit: ['xdev:device-model:edit'],
  modelTypeCreate: ['xdev:device-model-type:create'],
  modelTypeDelete: ['xdev:device-model-type:delete'],
  modelTypeEdit: ['xdev:device-model-type:edit'],
  relationCreate: ['xdev:device-model-parameter-group:create'],
  relationDelete: ['xdev:device-model-parameter-group:delete'],
} as const;

const GROUP_TYPE_OPTIONS = [
  'COMMUNICATION',
  'CONTROL',
  'ACQUISITION',
  'USER_DEFINITION',
  'INTEGRATION',
  'NO_CLASSIFY',
].map((value) => ({
  label: $t(`enum.deviceParameterGroup.groupType.${value}`),
  value,
}));
const SHARE_SCOPE_OPTIONS = [
  'NOT_SHARED',
  'DEVICE_GROUP_SHARED',
  'GLOBAL_SHARED',
].map((value) => ({
  label: $t(`enum.deviceParameterGroup.shareScope.${value}`),
  value,
}));
const MODEL_TYPE_USE_CASE_OPTIONS = [
  'NETWORK',
  'OFFICE',
  'OTHER',
  'PRODUCTION',
  'SECURITY',
].map((value) => ({
  label: $t(`enum.deviceModelType.useCase.${value}`),
  value,
}));
const VALUE_TYPE_OPTIONS = ['NUMBER', 'BOOL', 'STRING', 'JSON'].map(
  (value) => ({
    label: $t(`enum.deviceParameterItem.valueType.${value}`),
    value,
  }),
);
const CONSTRAINT_TYPE_OPTIONS = ['NONE', 'RANGE', 'LENGTH'].map((value) => ({
  label: $t(`enum.deviceParameterItem.constraintType.${value}`),
  value,
}));

const { hasAccessByCodes } = useAccess();
const userStore = useUserStore();
const currentTenantId = computed(() => userStore.userInfo?.tenantId ?? 0);

function pc(key: string, named?: Record<string, number | string>) {
  return $t(`page.parameterCenter.${key}`, named ?? {});
}

const leftCollapsed = ref(false);
const leftPaneRef = ref<HTMLElement>();
const topPanePercent = ref(55);
const modelLoading = ref(false);
const groupLoading = ref(false);
const itemLoading = ref(false);
const modelError = ref('');
const groupError = ref('');
const itemError = ref('');
const modelSearch = ref('');
const itemSearch = ref('');
const itemValueTypeFilter = ref<string>();
const itemConstraintTypeFilter = ref<string>();
const modelItems = ref<AdminDeviceModel[]>([]);
const modelTypeItems = ref<AdminDeviceModelType[]>([]);
const parameterGroups = ref<ParameterGroupRow[]>([]);
const parameterItems = ref<AdminDeviceParameterItem[]>([]);
const allModelRelations = ref<AdminDeviceModelParameterGroup[]>([]);
const selectedModelId = ref<number>();
const selectedParameterGroupId = ref<number>();
const itemPagination = reactive({ current: 1, pageSize: 20, total: 0 });

const bindOpen = ref(false);
const bindLoading = ref(false);
const bindSubmitting = ref(false);
const bindSearch = ref('');
const bindGroups = ref<ParameterGroupRow[]>([]);
const bindSelectedIds = ref<number[]>([]);

const groupDrawerOpen = ref(false);
const groupSubmitting = ref(false);
const groupFormRef = ref<FormInstance>();
const groupForm = reactive<Record<string, any>>({});
const groupEditingId = ref<number>();
const groupEditingModelCount = ref(0);
const groupDrawerTitle = computed(() =>
  groupEditingId.value ? pc('editGroup') : pc('createAndBindGroup'),
);

const copyDrawerOpen = ref(false);
const copySubmitting = ref(false);
const copyFormRef = ref<FormInstance>();
const copyForm = reactive({
  copyItems: true,
  groupCode: '',
  groupName: '',
  version: 1,
});

const associationOpen = ref(false);
const associationLoading = ref(false);
const associationTab = ref('models');
const associationModels = ref<AdminDeviceModel[]>([]);
const associationDevices = ref<AdminDevice[]>([]);
const associationGroups = ref<AssociationGroupRow[]>([]);

const entityModalOpen = ref(false);
const entitySubmitting = ref(false);
const entityKind = ref<'model' | 'type'>('type');
const entityEditingId = ref<number>();
const entityFormRef = ref<FormInstance>();
const entityForm = reactive<Record<string, any>>({});
const entityModalTitle = computed(() => {
  if (entityKind.value === 'type') {
    return entityEditingId.value ? pc('editDeviceType') : pc('addDeviceType');
  }
  return entityEditingId.value ? pc('editDeviceModel') : pc('addDeviceModel');
});

const itemDrawerOpen = ref(false);
const itemSubmitting = ref(false);
const itemFormRef = ref<FormInstance>();
const itemForm = reactive<Record<string, any>>({});
const constraintForm = reactive<{
  max?: number;
  maxLength?: number;
  min?: number;
}>({});
const itemMode = ref<'copy' | 'create' | 'detail' | 'edit'>('create');
const itemEditingId = ref<number>();
const itemDrawerTitle = computed(() => {
  const titles = {
    copy: pc('copyParameter'),
    create: pc('createParameter'),
    detail: pc('parameterDetail'),
    edit: pc('editParameter'),
  };
  return titles[itemMode.value];
});
const itemReadonly = computed(() => itemMode.value === 'detail');
const availableConstraintTypeOptions = computed(() => {
  if (itemForm.valueType === 'NUMBER') {
    return CONSTRAINT_TYPE_OPTIONS.filter(
      (item) => item.value === 'NONE' || item.value === 'RANGE',
    );
  }
  if (itemForm.valueType === 'STRING') {
    return CONSTRAINT_TYPE_OPTIONS.filter(
      (item) => item.value === 'NONE' || item.value === 'LENGTH',
    );
  }
  return CONSTRAINT_TYPE_OPTIONS.filter((item) => item.value === 'NONE');
});
const itemTargetGroupOptions = computed(() =>
  parameterGroups.value
    .filter(
      (group) =>
        Number(group.tenantId ?? 0) === Number(currentTenantId.value) &&
        canEditParameterGroup(group),
    )
    .map((group) => ({
      label: group.groupName,
      value: group.id,
    })),
);

const selectedModel = computed(() =>
  modelItems.value.find((item) => item.id === selectedModelId.value),
);

const selectedParameterGroup = computed(() =>
  parameterGroups.value.find(
    (item) => item.id === selectedParameterGroupId.value,
  ),
);

const selectedModelKeys = computed(() =>
  selectedModelId.value ? [`model:${selectedModelId.value}`] : [],
);

const filteredBindGroups = computed(() => {
  const keyword = bindSearch.value.trim().toLocaleLowerCase();
  return bindGroups.value.filter(
    (group) =>
      !keyword ||
      (group.groupName ?? '').toLocaleLowerCase().includes(keyword) ||
      (group.groupCode ?? '').toLocaleLowerCase().includes(keyword),
  );
});

const canBind = computed(
  () =>
    !!selectedModelId.value &&
    hasAccessByCodes([...PAGE_ACCESS.relationCreate]),
);
const boundGroupIds = computed(
  () => new Set(parameterGroups.value.map((group) => group.id)),
);

function canMutateGroup(group: AdminDeviceParameterGroup) {
  return Number(group.tenantId ?? 0) === Number(currentTenantId.value);
}

function canEditParameterGroup(group: AdminDeviceParameterGroup) {
  return canMutateGroup(group) && group.editable !== false;
}

function parameterGroupDisabledReason(group: AdminDeviceParameterGroup) {
  if (!canMutateGroup(group)) return pc('tenantMismatchReason');
  if (group.editable === false) return pc('readOnlyReason');
  return '';
}

function hasGroupMoreActions() {
  return (
    hasAccessByCodes([...PAGE_ACCESS.groupEdit]) ||
    hasAccessByCodes([...PAGE_ACCESS.groupCreate]) ||
    hasAccessByCodes([...PAGE_ACCESS.groupDelete])
  );
}

function groupMoreMenuItems(group: ParameterGroupRow): MenuProps['items'] {
  return [
    hasAccessByCodes([...PAGE_ACCESS.groupEdit])
      ? {
          disabled: !canMutateGroup(group),
          icon: () => h(IconifyIcon, { icon: 'lucide:pencil' }),
          key: 'edit',
          label: pc('editGroup'),
        }
      : null,
    hasAccessByCodes([...PAGE_ACCESS.groupCreate])
      ? {
          icon: () => h(IconifyIcon, { icon: 'lucide:copy' }),
          key: 'copy',
          label: pc('copyGroup'),
        }
      : null,
    hasAccessByCodes([...PAGE_ACCESS.groupDelete])
      ? {
          danger: true,
          disabled: !canMutateGroup(group),
          icon: () => h(IconifyIcon, { icon: 'lucide:trash-2' }),
          key: 'delete',
          label: pc('deleteGroup'),
        }
      : null,
  ].filter(Boolean) as MenuProps['items'];
}

function bindDisabledReason(group: Record<string, any>) {
  if (boundGroupIds.value.has(group.id)) return pc('bound');
  if (group.shareScope === 'NOT_SHARED' && group.modelCount > 0)
    return pc('notSharedAlreadyBound');
  return '';
}

const bindColumns: TableColumnsType<ParameterGroupRow> = [
  {
    dataIndex: 'groupName',
    key: 'groupName',
    title: pc('groupName'),
    width: 180,
  },
  {
    dataIndex: 'groupCode',
    key: 'groupCode',
    title: pc('groupCode'),
    width: 180,
  },
  {
    dataIndex: 'groupType',
    key: 'groupType',
    title: pc('groupType'),
    width: 130,
  },
  {
    dataIndex: 'shareScope',
    key: 'shareScope',
    title: pc('shareScope'),
    width: 130,
  },
  {
    dataIndex: 'itemCount',
    key: 'itemCount',
    title: pc('parameterName'),
    width: 90,
  },
  {
    dataIndex: 'modelCount',
    key: 'modelCount',
    title: pc('associatedModels'),
    width: 100,
  },
];

const associationModelColumns: TableColumnsType<AdminDeviceModel> = [
  { dataIndex: 'modelName', key: 'modelName', title: pc('modelName') },
  { dataIndex: 'description', key: 'description', title: pc('description') },
  { dataIndex: 'tenantId', key: 'tenantId', title: pc('tenant'), width: 100 },
];
const associationDeviceColumns: TableColumnsType<AdminDevice> = [
  { dataIndex: 'deviceCode', key: 'deviceCode', title: pc('deviceCode') },
  { dataIndex: 'name', key: 'name', title: pc('deviceName') },
  { dataIndex: 'serialNumber', key: 'serialNumber', title: pc('serialNumber') },
  { dataIndex: 'useStatus', key: 'useStatus', title: pc('status'), width: 100 },
];
const associationGroupColumns: TableColumnsType<AssociationGroupRow> = [
  { dataIndex: 'groupName', key: 'groupName', title: pc('deviceGroup') },
  { dataIndex: 'type', key: 'type', title: pc('groupType'), width: 130 },
  { dataIndex: 'tenantId', key: 'tenantId', title: pc('tenant'), width: 100 },
  {
    dataIndex: 'deviceCount',
    key: 'deviceCount',
    title: pc('relatedDevices'),
    width: 110,
  },
];

function resolveShareScope(value?: string) {
  return value ? $t(`enum.deviceParameterGroup.shareScope.${value}`) : '-';
}

function resolveConstraintSummary(item: AdminDeviceParameterItem) {
  if (item.constraintType === 'RANGE' && item.constraintConfig) {
    try {
      const config = JSON.parse(item.constraintConfig) as {
        max?: number;
        min?: number;
      };
      return `${config.min ?? '-'} ~ ${config.max ?? '-'}`;
    } catch {
      return item.constraintConfig;
    }
  }
  if (item.constraintType === 'LENGTH' && item.constraintConfig) {
    try {
      const config = JSON.parse(item.constraintConfig) as {
        maxLength?: number;
      };
      return pc('maxCharacters', { count: config.maxLength ?? '-' });
    } catch {
      return item.constraintConfig;
    }
  }
  return pc('none');
}

const modelTreeData = computed<TreeProps['treeData']>(() => {
  const keyword = modelSearch.value.trim().toLocaleLowerCase();
  return modelTypeItems.value.flatMap((type) => {
    if (typeof type.id !== 'number') return [];
    const typeName = type.modelTypeName ?? `#${type.id}`;
    const typeMatches = typeName.toLocaleLowerCase().includes(keyword);
    const children = modelItems.value
      .filter((model) => model.modelTypeId === type.id)
      .filter(
        (model) =>
          !keyword ||
          typeMatches ||
          (model.modelName ?? '').toLocaleLowerCase().includes(keyword),
      )
      .map((model) => ({
        dataKind: 'model',
        key: `model:${model.id}`,
        record: model,
        title: model.modelName ?? `#${model.id}`,
      }));
    if (keyword && !typeMatches && children.length === 0) return [];
    return [
      {
        children,
        dataKind: 'type',
        key: `type:${type.id}`,
        record: type,
        selectable: false,
        title: `${typeName} (${children.length})`,
      },
    ];
  });
});

const itemColumns: TableColumnsType<AdminDeviceParameterItem> = [
  {
    dataIndex: 'parameterName',
    key: 'parameterName',
    title: pc('parameterName'),
    width: 170,
  },
  {
    dataIndex: 'parameterKey',
    key: 'parameterKey',
    title: pc('parameterKey'),
    width: 160,
  },
  {
    dataIndex: 'valueType',
    key: 'valueType',
    title: pc('valueType'),
    width: 110,
  },
  {
    dataIndex: 'defaultValue',
    key: 'defaultValue',
    title: pc('defaultValue'),
    width: 160,
  },
  {
    dataIndex: 'constraintType',
    key: 'constraintType',
    title: pc('constraintType'),
    width: 180,
  },
  { dataIndex: 'unit', key: 'unit', title: pc('unit'), width: 90 },
  { dataIndex: 'required', key: 'required', title: pc('required'), width: 80 },
  {
    customCell: () => ({ class: 'parameter-remark-cell' }),
    dataIndex: 'remark',
    ellipsis: true,
    key: 'remark',
    minWidth: 260,
    title: pc('remark'),
    width: 260,
  },
  { fixed: 'right', key: 'action', title: pc('actions'), width: 190 },
];

async function loadModelTree() {
  modelLoading.value = true;
  modelError.value = '';
  const previousModelId = selectedModelId.value;
  try {
    const [types, models] = await Promise.all([
      listDeviceModelTypePage({ page: 1, pageSize: 500 }),
      listDeviceModelPage({ page: 1, pageSize: 1000 }),
    ]);
    modelTypeItems.value = types.items;
    modelItems.value = models.items;

    const initialModel =
      models.items.find((item) => item.id === previousModelId) ??
      models.items[0];
    selectedModelId.value = initialModel?.id;
    if (selectedModelId.value) await loadParameterGroups(selectedModelId.value);
  } catch (error) {
    modelError.value = (error as Error).message || pc('loadModelsFailed');
  } finally {
    modelLoading.value = false;
  }
}

async function loadParameterGroups(modelId: number) {
  groupLoading.value = true;
  groupError.value = '';
  const previousGroupId = selectedParameterGroupId.value;
  parameterGroups.value = [];
  selectedParameterGroupId.value = undefined;
  parameterItems.value = [];
  itemPagination.total = 0;
  try {
    const [modelRelations, allRelations, groups] = await Promise.all([
      listDeviceModelParameterGroupPage({
        modelId: String(modelId),
        page: 1,
        pageSize: 1000,
      }),
      listDeviceModelParameterGroupPage({ page: 1, pageSize: 5000 }),
      listDeviceParameterGroupPage({ page: 1, pageSize: 1000 }),
    ]);
    const boundIds = new Set(
      modelRelations.items
        .map((item) => item.parameterGroupId)
        .filter((id): id is number => typeof id === 'number'),
    );
    allModelRelations.value = allRelations.items;
    const relationByGroup = new Map(
      modelRelations.items
        .filter((item) => typeof item.parameterGroupId === 'number')
        .map((item) => [item.parameterGroupId as number, item]),
    );
    const modelCountByGroup = new Map<number, number>();
    for (const relation of allRelations.items) {
      if (typeof relation.parameterGroupId !== 'number') continue;
      modelCountByGroup.set(
        relation.parameterGroupId,
        (modelCountByGroup.get(relation.parameterGroupId) ?? 0) + 1,
      );
    }
    const boundGroups = groups.items.filter(
      (group) => typeof group.id === 'number' && boundIds.has(group.id),
    );
    const itemCounts = await Promise.all(
      boundGroups.map(async (group) => {
        const response = await listDeviceParameterItemPage({
          page: 1,
          pageSize: 1,
          parameterGroupId: String(group.id),
        });
        return response.total;
      }),
    );
    parameterGroups.value = boundGroups.map((group, index) => ({
      ...group,
      itemCount: itemCounts[index] ?? 0,
      modelCount: modelCountByGroup.get(group.id ?? 0) ?? 0,
      relationId: relationByGroup.get(group.id ?? 0)?.id,
    }));

    const initialGroup =
      parameterGroups.value.find((item) => item.id === previousGroupId) ??
      parameterGroups.value[0];
    if (initialGroup?.id) await selectParameterGroup(initialGroup.id);
  } catch (error) {
    groupError.value = (error as Error).message || pc('loadGroupsFailed');
  } finally {
    groupLoading.value = false;
  }
}

async function loadParameterItems(
  page = itemPagination.current,
  pageSize = itemPagination.pageSize,
) {
  if (!selectedParameterGroupId.value) {
    parameterItems.value = [];
    itemPagination.total = 0;
    return;
  }
  itemLoading.value = true;
  itemError.value = '';
  try {
    const response = await listDeviceParameterItemPage({
      page,
      pageSize,
      parameterGroupId: String(selectedParameterGroupId.value),
      parameterName: itemSearch.value || undefined,
      constraintType: itemConstraintTypeFilter.value,
      valueType: itemValueTypeFilter.value,
    });
    parameterItems.value = response.items;
    itemPagination.current = page;
    itemPagination.pageSize = pageSize;
    itemPagination.total = response.total;
  } catch (error) {
    itemError.value = (error as Error).message || pc('loadItemsFailed');
  } finally {
    itemLoading.value = false;
  }
}

async function handleModelSelect(keys: (number | string)[]) {
  const key = String(keys[0] ?? '');
  if (!key.startsWith('model:')) return;
  const modelId = Number(key.slice('model:'.length));
  if (!Number.isInteger(modelId) || modelId === selectedModelId.value) return;
  selectedModelId.value = modelId;
  await loadParameterGroups(modelId);
}

async function selectParameterGroup(groupId: number) {
  selectedParameterGroupId.value = groupId;
  itemPagination.current = 1;
  itemSearch.value = '';
  await loadParameterItems(1);
}

async function handleItemTableChange(pagination: TablePaginationConfig) {
  await loadParameterItems(pagination.current ?? 1, pagination.pageSize ?? 20);
}

async function handleItemSearch() {
  await loadParameterItems(1);
}

function canMutateTenant(tenantId?: number) {
  return Number(tenantId ?? 0) === Number(currentTenantId.value);
}

async function openCreateType() {
  entityKind.value = 'type';
  entityEditingId.value = undefined;
  Object.assign(entityForm, {
    modelTypeName: '',
    typeDesc: '',
    useCase: 'OTHER',
  });
  entityModalOpen.value = true;
  await nextTick();
  entityFormRef.value?.clearValidate();
}

async function openEditType(record: AdminDeviceModelType) {
  if (!record.id) return;
  const data = await getDeviceModelTypeById(record.id);
  if (!data) return;
  entityKind.value = 'type';
  entityEditingId.value = record.id;
  Object.assign(entityForm, data);
  entityModalOpen.value = true;
  await nextTick();
  entityFormRef.value?.clearValidate();
}

async function openCreateModel(typeId?: number) {
  entityKind.value = 'model';
  entityEditingId.value = undefined;
  Object.assign(entityForm, {
    description: '',
    modelName: '',
    modelTypeId: typeId ?? selectedModel.value?.modelTypeId,
    remark: '',
  });
  entityModalOpen.value = true;
  await nextTick();
  entityFormRef.value?.clearValidate();
}

async function openEditModel(record: AdminDeviceModel) {
  if (!record.id) return;
  const data = await getDeviceModelById(record.id);
  if (!data) return;
  entityKind.value = 'model';
  entityEditingId.value = record.id;
  Object.assign(entityForm, data);
  entityModalOpen.value = true;
  await nextTick();
  entityFormRef.value?.clearValidate();
}

async function handleEntitySubmit() {
  await entityFormRef.value?.validate();
  entitySubmitting.value = true;
  try {
    if (entityKind.value === 'type') {
      await (entityEditingId.value
        ? updateDeviceModelType(entityEditingId.value, entityForm)
        : createDeviceModelType({
            ...entityForm,
            tenantId: currentTenantId.value,
          }));
    } else if (entityEditingId.value) {
      await updateDeviceModel(entityEditingId.value, entityForm);
    } else {
      await createDeviceModel({
        ...entityForm,
        tenantId: currentTenantId.value,
      });
    }
    const modelName =
      entityKind.value === 'model' ? entityForm.modelName : undefined;
    entityModalOpen.value = false;
    await loadModelTree();
    if (modelName) {
      const model = modelItems.value.find(
        (item) => item.modelName === modelName,
      );
      if (model?.id && model.id !== selectedModelId.value) {
        selectedModelId.value = model.id;
        await loadParameterGroups(model.id);
      }
    }
    message.success(pc('saved'));
  } finally {
    entitySubmitting.value = false;
  }
}

async function handleDeleteType(record: AdminDeviceModelType) {
  if (!record.id) return;
  const models = await listDeviceModelPage({
    modelTypeId: String(record.id),
    page: 1,
    pageSize: 1,
  });
  if (models.total > 0) {
    Modal.warning({
      content: pc('typeHasModels', { count: models.total }),
      title: pc('cannotDeleteType'),
    });
    return;
  }
  Modal.confirm({
    content: pc('confirmDeleteType', { name: record.modelTypeName ?? '' }),
    okButtonProps: { danger: true },
    title: pc('deleteDeviceType'),
    async onOk() {
      await deleteDeviceModelType(record.id as number);
      await loadModelTree();
      message.success(pc('typeDeleted'));
    },
  });
}

async function handleDeleteModel(record: AdminDeviceModel) {
  if (!record.id) return;
  const [devices, relations] = await Promise.all([
    listDevicePage({ modelId: String(record.id), page: 1, pageSize: 1 }),
    listDeviceModelParameterGroupPage({
      modelId: String(record.id),
      page: 1,
      pageSize: 1,
    }),
  ]);
  if (devices.total > 0 || relations.total > 0) {
    Modal.warning({
      content: pc('modelHasReferences', {
        devices: devices.total,
        groups: relations.total,
      }),
      title: pc('cannotDeleteModel'),
    });
    return;
  }
  Modal.confirm({
    content: pc('confirmDeleteModel', { name: record.modelName ?? '' }),
    okButtonProps: { danger: true },
    title: pc('deleteDeviceModel'),
    async onOk() {
      await deleteDeviceModel(record.id as number);
      if (selectedModelId.value === record.id)
        selectedModelId.value = undefined;
      await loadModelTree();
      message.success(pc('modelDeleted'));
    },
  });
}

function resetItemForm() {
  Object.assign(itemForm, {
    constraintConfig: '',
    constraintType: 'NONE',
    defaultValue: '',
    parameterGroupId: selectedParameterGroupId.value,
    parameterKey: '',
    parameterName: '',
    remark: '',
    required: false,
    unit: '',
    valueType: 'STRING',
  });
  Object.assign(constraintForm, {
    max: undefined,
    maxLength: undefined,
    min: undefined,
  });
}

function loadConstraintForm(config?: string) {
  Object.assign(constraintForm, {
    max: undefined,
    maxLength: undefined,
    min: undefined,
  });
  if (!config) return;
  try {
    const parsed = JSON.parse(config) as Record<string, unknown>;
    constraintForm.min =
      typeof parsed.min === 'number' ? parsed.min : undefined;
    constraintForm.max =
      typeof parsed.max === 'number' ? parsed.max : undefined;
    constraintForm.maxLength =
      typeof parsed.maxLength === 'number' ? parsed.maxLength : undefined;
  } catch {
    // Invalid stored configuration remains visible through the raw detail data.
  }
}

function handleValueTypeChange() {
  const allowed = new Set(
    availableConstraintTypeOptions.value.map((item) => item.value),
  );
  if (!allowed.has(itemForm.constraintType)) itemForm.constraintType = 'NONE';
  if (
    itemForm.valueType === 'BOOL' &&
    !['true', 'false'].includes(itemForm.defaultValue)
  ) {
    itemForm.defaultValue = 'false';
  }
}

async function openCreateItem() {
  itemMode.value = 'create';
  itemEditingId.value = undefined;
  resetItemForm();
  itemDrawerOpen.value = true;
  await nextTick();
  itemFormRef.value?.clearValidate();
}

async function openItem(
  record: AdminDeviceParameterItem,
  mode: 'copy' | 'detail' | 'edit',
) {
  if (!record.id) return;
  const data = await getDeviceParameterItemById(record.id);
  if (!data) return;
  itemMode.value = mode;
  itemEditingId.value = mode === 'edit' ? record.id : undefined;
  resetItemForm();
  Object.assign(itemForm, data);
  loadConstraintForm(data.constraintConfig);
  if (mode === 'copy') {
    itemForm.id = undefined;
    itemForm.parameterKey = `${data.parameterKey ?? 'parameter'}_copy`;
    itemForm.parameterName = pc('copySuffix', {
      name: data.parameterName ?? pc('parameter'),
    });
  }
  itemDrawerOpen.value = true;
  await nextTick();
  itemFormRef.value?.clearValidate();
}

async function validateItemForm() {
  await itemFormRef.value?.validate();
  const matches = await listDeviceParameterItemPage({
    page: 1,
    pageSize: 100,
    parameterGroupId: String(itemForm.parameterGroupId),
    parameterKey: itemForm.parameterKey,
  });
  const duplicate = matches.items.some(
    (item) =>
      item.parameterKey === itemForm.parameterKey &&
      item.id !== itemEditingId.value,
  );
  if (duplicate) throw new Error(pc('keyDuplicate'));

  if (itemForm.constraintType === 'RANGE') {
    if (
      !Number.isFinite(constraintForm.min) ||
      !Number.isFinite(constraintForm.max)
    ) {
      throw new TypeError(pc('invalidRange'));
    }
    if ((constraintForm.min as number) > (constraintForm.max as number)) {
      throw new Error(pc('rangeOrder'));
    }
    itemForm.constraintConfig = JSON.stringify({
      max: constraintForm.max,
      min: constraintForm.min,
    });
  } else if (itemForm.constraintType === 'LENGTH') {
    if (
      !Number.isInteger(constraintForm.maxLength) ||
      (constraintForm.maxLength ?? 0) <= 0
    ) {
      throw new Error(pc('invalidMaxLength'));
    }
    itemForm.constraintConfig = JSON.stringify({
      maxLength: constraintForm.maxLength,
    });
  } else {
    itemForm.constraintConfig = '';
  }

  if (itemForm.valueType === 'NUMBER' && itemForm.defaultValue !== '') {
    const value = Number(itemForm.defaultValue);
    if (!Number.isFinite(value)) throw new Error(pc('invalidNumber'));
    if (
      itemForm.constraintType === 'RANGE' &&
      (value < (constraintForm.min as number) ||
        value > (constraintForm.max as number))
    ) {
      throw new Error(pc('defaultOutOfRange'));
    }
  }
  if (
    itemForm.valueType === 'BOOL' &&
    !['true', 'false', ''].includes(itemForm.defaultValue)
  ) {
    throw new Error(pc('invalidBoolean'));
  }
  if (
    itemForm.valueType === 'STRING' &&
    itemForm.constraintType === 'LENGTH' &&
    (itemForm.defaultValue?.length ?? 0) > (constraintForm.maxLength as number)
  ) {
    throw new Error(pc('defaultTooLong'));
  }
  if (itemForm.valueType === 'JSON' && itemForm.defaultValue) {
    try {
      JSON.parse(itemForm.defaultValue);
    } catch {
      throw new Error(pc('invalidJson'));
    }
  }
}

async function handleItemSubmit(continueCreate = false) {
  if (itemReadonly.value) {
    itemDrawerOpen.value = false;
    return;
  }
  await validateItemForm();
  itemSubmitting.value = true;
  try {
    await (itemEditingId.value
      ? updateDeviceParameterItem(itemEditingId.value, itemForm)
      : createDeviceParameterItem({
          ...itemForm,
          tenantId: currentTenantId.value,
        }));
    await loadParameterItems(1);
    if (selectedModelId.value) await loadParameterGroups(selectedModelId.value);
    message.success(itemEditingId.value ? pc('updated') : pc('created'));
    if (continueCreate && !itemEditingId.value) {
      resetItemForm();
      await nextTick();
      itemFormRef.value?.clearValidate();
    } else {
      itemDrawerOpen.value = false;
    }
  } finally {
    itemSubmitting.value = false;
  }
}

function handleDeleteItem(record: AdminDeviceParameterItem) {
  if (!record.id) return;
  Modal.confirm({
    content: pc('confirmDeleteParameter', {
      count: selectedParameterGroup.value?.modelCount ?? 0,
      key: record.parameterKey ?? '',
      name: record.parameterName ?? '',
    }),
    okButtonProps: { danger: true },
    title: pc('delete'),
    async onOk() {
      await deleteDeviceParameterItem(record.id as number);
      await loadParameterItems();
      if (selectedModelId.value)
        await loadParameterGroups(selectedModelId.value);
      message.success(pc('deleted'));
    },
  });
}

function resetGroupForm() {
  Object.assign(groupForm, {
    description: '',
    editable: true,
    groupCode: '',
    groupName: '',
    groupType: 'NO_CLASSIFY',
    shareScope: 'NOT_SHARED',
    version: 1,
  });
}

async function findCreatedGroup(groupCode: string) {
  const result = await listDeviceParameterGroupPage({
    groupCode,
    page: 1,
    pageSize: 100,
  });
  return result.items.find((item) => item.groupCode === groupCode);
}

async function bindGroupToSelectedModel(parameterGroupId: number) {
  if (!selectedModelId.value) throw new Error(pc('selectModelRequired'));
  await createDeviceModelParameterGroup({
    modelId: selectedModelId.value,
    parameterGroupId,
    tenantId: selectedModel.value?.tenantId,
  });
}

async function loadGroupSummaries(groups: AdminDeviceParameterGroup[]) {
  const itemCounts = await Promise.all(
    groups.map(async (group) => {
      if (!group.id) return 0;
      const result = await listDeviceParameterItemPage({
        page: 1,
        pageSize: 1,
        parameterGroupId: String(group.id),
      });
      return result.total;
    }),
  );
  const counts = new Map<number, number>();
  for (const relation of allModelRelations.value) {
    if (!relation.parameterGroupId) continue;
    counts.set(
      relation.parameterGroupId,
      (counts.get(relation.parameterGroupId) ?? 0) + 1,
    );
  }
  return groups.map((group, index) => ({
    ...group,
    itemCount: itemCounts[index] ?? 0,
    modelCount: counts.get(group.id ?? 0) ?? 0,
  }));
}

async function openBindModal() {
  if (!selectedModelId.value) return;
  bindOpen.value = true;
  bindLoading.value = true;
  bindSearch.value = '';
  bindSelectedIds.value = [];
  try {
    const [groups, relations] = await Promise.all([
      listDeviceParameterGroupPage({ page: 1, pageSize: 1000 }),
      listDeviceModelParameterGroupPage({ page: 1, pageSize: 5000 }),
    ]);
    allModelRelations.value = relations.items;
    bindGroups.value = await loadGroupSummaries(groups.items);
  } catch (error) {
    message.error((error as Error).message || pc('loadGroupLibraryFailed'));
  } finally {
    bindLoading.value = false;
  }
}

async function handleBindGroups() {
  if (!selectedModelId.value || bindSelectedIds.value.length === 0) return;
  const bindableIds = new Set(
    bindGroups.value
      .filter((group) => !bindDisabledReason(group))
      .map((group) => group.id),
  );
  const selectedIds = bindSelectedIds.value.filter((id) => bindableIds.has(id));
  if (selectedIds.length === 0) return;
  bindSubmitting.value = true;
  try {
    const results = await Promise.allSettled(
      selectedIds.map((id) => bindGroupToSelectedModel(id)),
    );
    const failed = results.filter((result) => result.status === 'rejected');
    const succeeded = results.length - failed.length;
    if (failed.length > 0) {
      message.warning(
        pc('partialBind', { failed: failed.length, success: succeeded }),
      );
    } else {
      message.success(pc('bindSuccess', { count: succeeded }));
      bindOpen.value = false;
    }
    await loadParameterGroups(selectedModelId.value);
  } finally {
    bindSubmitting.value = false;
  }
}

function handleBindSelectionChange(keys: (number | string)[]) {
  bindSelectedIds.value = keys.map(Number);
}

async function openCreateAndBind() {
  groupEditingId.value = undefined;
  groupEditingModelCount.value = 0;
  resetGroupForm();
  groupDrawerOpen.value = true;
  await nextTick();
  groupFormRef.value?.clearValidate();
}

async function showEditGroupDrawer(group: ParameterGroupRow) {
  if (!group.id) return;
  const data = await getDeviceParameterGroupById(group.id);
  if (!data) {
    message.warning(pc('groupMissing'));
    return;
  }
  groupEditingId.value = group.id;
  groupEditingModelCount.value = group.modelCount;
  resetGroupForm();
  Object.assign(groupForm, data);
  groupDrawerOpen.value = true;
  await nextTick();
  groupFormRef.value?.clearValidate();
}

function openEditGroup(group: ParameterGroupRow) {
  if (group.modelCount <= 1) {
    void showEditGroupDrawer(group);
    return;
  }
  Modal.confirm({
    content: pc('sharedEditImpact', { count: group.modelCount }),
    okText: pc('continueEdit'),
    title: pc('confirmSharedEdit'),
    onOk: () => showEditGroupDrawer(group),
  });
}

async function handleCreateAndBind() {
  await groupFormRef.value?.validate();
  groupSubmitting.value = true;
  try {
    if (groupEditingId.value) {
      const editingId = groupEditingId.value;
      await updateDeviceParameterGroup(editingId, groupForm);
      groupDrawerOpen.value = false;
      await loadParameterGroups(selectedModelId.value as number);
      await selectParameterGroup(editingId);
      message.success(pc('groupUpdated'));
      return;
    }
    await createDeviceParameterGroup({
      ...groupForm,
      tenantId: selectedModel.value?.tenantId,
    });
    const created = await findCreatedGroup(groupForm.groupCode);
    if (!created?.id) throw new Error(pc('createdGroupMissing'));
    await bindGroupToSelectedModel(created.id);
    groupDrawerOpen.value = false;
    await loadParameterGroups(selectedModelId.value as number);
    await selectParameterGroup(created.id);
    message.success(pc('groupCreatedBound'));
  } finally {
    groupSubmitting.value = false;
  }
}

async function handleUnbind(group: ParameterGroupRow) {
  if (!group.relationId || !selectedModelId.value) return;
  await deleteDeviceModelParameterGroup(group.relationId);
  message.success(
    pc('unbindSuccess', {
      group: group.groupName ?? group.groupCode ?? '',
      model: selectedModel.value?.modelName ?? pc('currentModel'),
    }),
  );
  await loadParameterGroups(selectedModelId.value);
}

async function openCopyDrawer(group: ParameterGroupRow) {
  selectedParameterGroupId.value = group.id;
  Object.assign(copyForm, {
    copyItems: true,
    groupCode: `${group.groupCode ?? 'group'}_copy`,
    groupName: pc('copySuffix', {
      name: group.groupName ?? group.groupCode ?? pc('groupName'),
    }),
    version: group.version ?? 1,
  });
  copyDrawerOpen.value = true;
  await nextTick();
  copyFormRef.value?.clearValidate();
}

async function handleCopyGroup() {
  const source = selectedParameterGroup.value;
  if (!source?.id || !selectedModelId.value) return;
  await copyFormRef.value?.validate();
  copySubmitting.value = true;
  let createdId: number | undefined;
  try {
    await createDeviceParameterGroup({
      ...source,
      id: undefined,
      groupCode: copyForm.groupCode,
      groupName: copyForm.groupName,
      tenantId: selectedModel.value?.tenantId,
      version: copyForm.version,
    });
    const created = await findCreatedGroup(copyForm.groupCode);
    createdId = created?.id;
    if (!createdId) throw new Error(pc('copyCreatedMissing'));
    if (copyForm.copyItems) {
      const sourceItems = await listDeviceParameterItemPage({
        page: 1,
        pageSize: 5000,
        parameterGroupId: String(source.id),
      });
      for (const item of sourceItems.items) {
        await createDeviceParameterItem({
          ...item,
          id: undefined,
          parameterGroupId: createdId,
        });
      }
    }
    await bindGroupToSelectedModel(createdId);
    copyDrawerOpen.value = false;
    await loadParameterGroups(selectedModelId.value);
    await selectParameterGroup(createdId);
    message.success(pc('groupCopiedBound'));
  } catch (error) {
    if (createdId) {
      message.error(pc('copyPartial', { id: createdId }));
    }
    throw error;
  } finally {
    copySubmitting.value = false;
  }
}

async function openAssociations(group: ParameterGroupRow) {
  if (!group.id) return;
  selectedParameterGroupId.value = group.id;
  associationOpen.value = true;
  associationLoading.value = true;
  associationTab.value = 'models';
  try {
    const relations = await listDeviceModelParameterGroupPage({
      page: 1,
      pageSize: 5000,
      parameterGroupId: String(group.id),
    });
    const modelIds = new Set(
      relations.items
        .map((item) => item.modelId)
        .filter((id): id is number => !!id),
    );
    associationModels.value = modelItems.value.filter(
      (model) => !!model.id && modelIds.has(model.id),
    );

    const deviceResults = await Promise.all(
      [...modelIds].map((modelId) =>
        listDevicePage({ modelId: String(modelId), page: 1, pageSize: 1000 }),
      ),
    );
    associationDevices.value = deviceResults.flatMap((result) => result.items);
    const deviceIds = associationDevices.value
      .map((item) => item.id)
      .filter((id): id is number => !!id);
    const relationResults = await Promise.all(
      deviceIds.map((deviceId) =>
        listDeviceGroupDevicePage({
          deviceId: String(deviceId),
          page: 1,
          pageSize: 1000,
        }),
      ),
    );
    const groupRelations = relationResults.flatMap((result) => result.items);
    const countByGroup = new Map<number, number>();
    for (const relation of groupRelations) {
      if (!relation.groupId) continue;
      countByGroup.set(
        relation.groupId,
        (countByGroup.get(relation.groupId) ?? 0) + 1,
      );
    }
    const groups = await listDeviceGroupPage({ page: 1, pageSize: 1000 });
    associationGroups.value = groups.items
      .filter((item) => !!item.id && countByGroup.has(item.id))
      .map((item) => ({
        ...item,
        deviceCount: countByGroup.get(item.id ?? 0) ?? 0,
      }));
  } catch (error) {
    message.error((error as Error).message || pc('loadAssociationsFailed'));
  } finally {
    associationLoading.value = false;
  }
}

async function handleDeleteGroup(group: ParameterGroupRow) {
  if (!group.id) return;
  if (group.modelCount > 1) {
    Modal.warning({
      content: pc('groupStillReferenced', { count: group.modelCount }),
      title: pc('cannotDeleteGroup'),
    });
    return;
  }
  Modal.confirm({
    content: pc('confirmDeleteGroup', {
      count: group.itemCount,
      name: group.groupName ?? group.groupCode ?? '',
    }),
    okButtonProps: { danger: true },
    okText: pc('confirmDelete'),
    title: pc('deleteGroup'),
    async onOk() {
      await deleteDeviceParameterGroup(group.id as number);
      message.success(pc('groupDeleted'));
      if (selectedModelId.value)
        await loadParameterGroups(selectedModelId.value);
    },
  });
}

function handleGroupMoreAction(key: string, group: ParameterGroupRow) {
  if (key === 'edit') openEditGroup(group);
  else if (key === 'copy') void openCopyDrawer(group);
  else if (key === 'delete') void handleDeleteGroup(group);
}

function beginVerticalResize(event: PointerEvent) {
  event.preventDefault();
  const move = (moveEvent: PointerEvent) => {
    const element = leftPaneRef.value;
    if (!element) return;
    const rect = element.getBoundingClientRect();
    const percent = ((moveEvent.clientY - rect.top) / rect.height) * 100;
    topPanePercent.value = Math.min(72, Math.max(30, percent));
  };
  const stop = () => {
    window.removeEventListener('pointermove', move);
    window.removeEventListener('pointerup', stop);
  };
  window.addEventListener('pointermove', move);
  window.addEventListener('pointerup', stop, { once: true });
}

onMounted(loadModelTree);
</script>

<template>
  <Page auto-content-height :title="pc('title')">
    <div
      :class="[
        'parameter-management',
        { 'parameter-management--collapsed': leftCollapsed },
      ]"
    >
      <aside ref="leftPaneRef" class="context-workbench">
        <Button
          class="collapse-button"
          size="small"
          type="text"
          :title="leftCollapsed ? pc('expandContext') : pc('collapseContext')"
          @click="leftCollapsed = !leftCollapsed"
        >
          <template #icon>
            <IconifyIcon
              :icon="
                leftCollapsed
                  ? 'lucide:panel-left-open'
                  : 'lucide:panel-left-close'
              "
            />
          </template>
        </Button>

        <template v-if="!leftCollapsed">
          <section
            class="context-section"
            :style="{ height: `${topPanePercent}%` }"
          >
            <div class="section-header">
              <div>
                <div class="section-title">{{ pc('deviceTypeAndModel') }}</div>
                <div class="section-subtitle">
                  {{ pc('modelCount', { count: modelItems.length }) }}
                </div>
              </div>
              <Space>
                <Button
                  v-if="hasAccessByCodes([...PAGE_ACCESS.modelTypeCreate])"
                  size="small"
                  type="text"
                  :title="pc('addDeviceType')"
                  @click="openCreateType"
                >
                  <template #icon
                    ><IconifyIcon icon="lucide:folder-plus"
                  /></template>
                </Button>
                <Button
                  v-if="hasAccessByCodes([...PAGE_ACCESS.modelCreate])"
                  size="small"
                  type="text"
                  :title="pc('addDeviceModel')"
                  @click="openCreateModel()"
                >
                  <template #icon><IconifyIcon icon="lucide:plus" /></template>
                </Button>
              </Space>
            </div>
            <Input
              v-model:value="modelSearch"
              allow-clear
              class="context-search"
              :placeholder="pc('searchTypeOrModel')"
            >
              <template #prefix><IconifyIcon icon="lucide:search" /></template>
            </Input>
            <Alert
              v-if="modelError"
              :message="modelError"
              show-icon
              type="error"
            />
            <div v-else class="tree-scroll">
              <Tree
                default-expand-all
                :loading="modelLoading"
                :selected-keys="selectedModelKeys"
                :tree-data="modelTreeData"
                @select="handleModelSelect"
              >
                <template #title="{ dataKind, record, title }">
                  <div class="model-tree-node">
                    <span class="text-ellipsis">{{ title }}</span>
                    <span class="model-tree-node__actions">
                      <Button
                        v-if="
                          dataKind === 'type' &&
                          hasAccessByCodes([...PAGE_ACCESS.modelCreate])
                        "
                        size="small"
                        type="text"
                        :title="pc('addModelUnderType')"
                        @click.stop="openCreateModel(record.id)"
                      >
                        <template #icon
                          ><IconifyIcon icon="lucide:plus"
                        /></template>
                      </Button>
                      <Button
                        v-if="
                          dataKind === 'type' &&
                          hasAccessByCodes([...PAGE_ACCESS.modelTypeEdit])
                        "
                        :disabled="!canMutateTenant(record.tenantId)"
                        size="small"
                        type="text"
                        :title="pc('editDeviceType')"
                        @click.stop="openEditType(record)"
                      >
                        <template #icon
                          ><IconifyIcon icon="lucide:pencil"
                        /></template>
                      </Button>
                      <Button
                        v-if="
                          dataKind === 'model' &&
                          hasAccessByCodes([...PAGE_ACCESS.modelEdit])
                        "
                        :disabled="!canMutateTenant(record.tenantId)"
                        size="small"
                        type="text"
                        :title="pc('editDeviceModel')"
                        @click.stop="openEditModel(record)"
                      >
                        <template #icon
                          ><IconifyIcon icon="lucide:pencil"
                        /></template>
                      </Button>
                      <Button
                        v-if="
                          dataKind === 'type' &&
                          hasAccessByCodes([...PAGE_ACCESS.modelTypeDelete])
                        "
                        :disabled="!canMutateTenant(record.tenantId)"
                        danger
                        size="small"
                        type="text"
                        :title="pc('deleteDeviceType')"
                        @click.stop="handleDeleteType(record)"
                      >
                        <template #icon
                          ><IconifyIcon icon="lucide:trash-2"
                        /></template>
                      </Button>
                      <Button
                        v-if="
                          dataKind === 'model' &&
                          hasAccessByCodes([...PAGE_ACCESS.modelDelete])
                        "
                        :disabled="!canMutateTenant(record.tenantId)"
                        danger
                        size="small"
                        type="text"
                        :title="pc('deleteDeviceModel')"
                        @click.stop="handleDeleteModel(record)"
                      >
                        <template #icon
                          ><IconifyIcon icon="lucide:trash-2"
                        /></template>
                      </Button>
                    </span>
                  </div>
                </template>
              </Tree>
            </div>
          </section>

          <div
            class="vertical-resizer"
            :title="pc('resizeContext')"
            @pointerdown="beginVerticalResize"
          >
            <span />
          </div>

          <section class="context-section group-section">
            <div class="section-header">
              <div class="min-w-0">
                <div class="section-title">{{ pc('currentModelGroups') }}</div>
                <div class="section-subtitle text-ellipsis">
                  {{ selectedModel?.modelName ?? pc('noModelSelected') }}
                </div>
              </div>
              <Space>
                <Button
                  :disabled="!canBind"
                  size="small"
                  type="text"
                  :title="pc('bindExistingGroup')"
                  @click="openBindModal"
                >
                  <template #icon><IconifyIcon icon="lucide:link" /></template>
                </Button>
                <Button
                  v-if="hasAccessByCodes([...PAGE_ACCESS.groupCreate])"
                  :disabled="!selectedModelId || !canBind"
                  size="small"
                  type="text"
                  :title="pc('createAndBindGroup')"
                  @click="openCreateAndBind"
                >
                  <template #icon><IconifyIcon icon="lucide:plus" /></template>
                </Button>
              </Space>
            </div>
            <Alert
              v-if="groupError"
              :message="groupError"
              show-icon
              type="error"
            />
            <div v-else-if="groupLoading" class="section-placeholder">
              {{ pc('loadingGroups') }}
            </div>
            <div v-else-if="parameterGroups.length > 0" class="group-list">
              <div
                v-for="group in parameterGroups"
                :key="group.id"
                :class="[
                  'group-item',
                  {
                    'group-item--selected':
                      group.id === selectedParameterGroupId,
                  },
                ]"
              >
                <button
                  class="group-item__content"
                  type="button"
                  @click="group.id && selectParameterGroup(group.id)"
                >
                  <span class="group-item__main">
                    <span class="group-item__name">{{
                      group.groupName ?? group.groupCode
                    }}</span>
                    <span class="group-item__count">{{
                      pc('itemCount', { count: group.itemCount })
                    }}</span>
                  </span>
                  <span class="group-item__meta">
                    <span>{{ group.groupCode }}</span>
                    <span>{{ resolveShareScope(group.shareScope) }}</span>
                    <Tag v-if="group.modelCount > 1" color="blue">{{
                      pc('modelsSummary', { count: group.modelCount })
                    }}</Tag>
                    <span>v{{ group.version ?? 1 }}</span>
                  </span>
                </button>
                <div class="group-item__actions">
                  <Button
                    size="small"
                    type="text"
                    :title="pc('association')"
                    @click.stop="openAssociations(group)"
                  >
                    <template #icon
                      ><IconifyIcon icon="lucide:network"
                    /></template>
                  </Button>
                  <Popconfirm
                    v-if="hasAccessByCodes([...PAGE_ACCESS.relationDelete])"
                    :title="pc('unbind')"
                    @confirm="handleUnbind(group)"
                  >
                    <Button size="small" type="text" :title="pc('unbind')">
                      <template #icon
                        ><IconifyIcon icon="lucide:unlink"
                      /></template>
                    </Button>
                  </Popconfirm>
                  <Dropdown
                    v-if="hasGroupMoreActions()"
                    placement="bottomRight"
                    trigger="click"
                  >
                    <Button size="small" type="text" :title="pc('moreActions')">
                      <template #icon
                        ><IconifyIcon icon="lucide:ellipsis"
                      /></template>
                    </Button>
                    <template #overlay>
                      <Menu
                        :items="groupMoreMenuItems(group)"
                        @click="
                          ({ key }) => handleGroupMoreAction(String(key), group)
                        "
                      />
                    </template>
                  </Dropdown>
                </div>
              </div>
            </div>
            <Empty
              v-else
              :description="
                selectedModelId ? pc('noBoundGroups') : pc('selectModelFirst')
              "
            />
          </section>
        </template>

        <div v-else class="collapsed-context">
          <IconifyIcon icon="lucide:cpu" />
          <span>{{ selectedModel?.modelName ?? pc('deviceModel') }}</span>
          <IconifyIcon icon="lucide:sliders-horizontal" />
          <span>{{
            selectedParameterGroup?.groupName ?? pc('groupName')
          }}</span>
        </div>
      </aside>

      <main class="parameter-workspace">
        <div class="workspace-header">
          <div class="min-w-0">
            <div class="workspace-eyebrow">{{ pc('currentContext') }}</div>
            <div class="workspace-title">
              {{ selectedModel?.modelName ?? pc('selectDeviceModel') }}
              <span v-if="selectedParameterGroup"
                >/ {{ selectedParameterGroup.groupName }}</span
              >
            </div>
            <div v-if="selectedParameterGroup" class="workspace-meta">
              {{ selectedParameterGroup.groupCode }} ·
              {{ selectedParameterGroup.groupType }} ·
              {{
                pc('modelReferences', {
                  count: selectedParameterGroup.modelCount,
                })
              }}
            </div>
          </div>
          <Space>
            <Button
              :disabled="!selectedParameterGroup"
              :title="pc('association')"
              @click="
                selectedParameterGroup &&
                openAssociations(selectedParameterGroup)
              "
            >
              <template #icon><IconifyIcon icon="lucide:network" /></template>
              {{ pc('association') }}
            </Button>
            <Button
              v-if="hasAccessByCodes([...PAGE_ACCESS.itemCreate])"
              :disabled="
                !selectedParameterGroup ||
                !canEditParameterGroup(selectedParameterGroup)
              "
              :title="
                selectedParameterGroup &&
                !canEditParameterGroup(selectedParameterGroup)
                  ? parameterGroupDisabledReason(selectedParameterGroup)
                  : pc('addParameter')
              "
              type="primary"
              @click="openCreateItem"
            >
              <template #icon><IconifyIcon icon="lucide:plus" /></template>
              {{ pc('addParameter') }}
            </Button>
          </Space>
        </div>

        <template v-if="selectedParameterGroup">
          <div class="parameter-toolbar">
            <Input
              v-model:value="itemSearch"
              allow-clear
              :placeholder="pc('searchParameter')"
              @press-enter="handleItemSearch"
            >
              <template #prefix><IconifyIcon icon="lucide:search" /></template>
            </Input>
            <Select
              v-model:value="itemValueTypeFilter"
              allow-clear
              :options="VALUE_TYPE_OPTIONS"
              :placeholder="pc('valueType')"
              @change="handleItemSearch"
            />
            <Select
              v-model:value="itemConstraintTypeFilter"
              allow-clear
              :options="CONSTRAINT_TYPE_OPTIONS"
              :placeholder="pc('constraintType')"
              @change="handleItemSearch"
            />
            <Button @click="handleItemSearch">{{ pc('query') }}</Button>
          </div>
          <Alert v-if="itemError" :message="itemError" show-icon type="error" />
          <Table
            class="parameter-table"
            :columns="itemColumns"
            :data-source="parameterItems"
            :loading="itemLoading"
            :pagination="itemPagination"
            :scroll="{ x: 1050, y: 'calc(100vh - 370px)' }"
            row-key="id"
            size="middle"
            @change="handleItemTableChange"
          >
            <template #bodyCell="{ column, record, text }">
              <template v-if="column.key === 'required'">
                <Tag :color="text ? 'green' : 'default'">{{
                  text ? pc('yes') : pc('no')
                }}</Tag>
              </template>
              <template v-else-if="column.key === 'valueType'">
                {{ $t(`enum.deviceParameterItem.valueType.${text}`) }}
              </template>
              <template v-else-if="column.key === 'constraintType'">
                {{ resolveConstraintSummary(record) }}
              </template>
              <template v-else-if="column.key === 'action'">
                <Space>
                  <Button
                    size="small"
                    type="link"
                    @click="openItem(record, 'detail')"
                    >{{ pc('detail') }}</Button
                  >
                  <Button
                    v-if="hasAccessByCodes([...PAGE_ACCESS.itemEdit])"
                    :disabled="!canEditParameterGroup(selectedParameterGroup)"
                    :title="
                      parameterGroupDisabledReason(selectedParameterGroup)
                    "
                    size="small"
                    type="link"
                    @click="openItem(record, 'edit')"
                    >{{ pc('edit') }}</Button
                  >
                  <Button
                    v-if="hasAccessByCodes([...PAGE_ACCESS.itemCreate])"
                    size="small"
                    type="link"
                    @click="openItem(record, 'copy')"
                    >{{ pc('copy') }}</Button
                  >
                  <Button
                    v-if="hasAccessByCodes([...PAGE_ACCESS.itemDelete])"
                    :disabled="!canEditParameterGroup(selectedParameterGroup)"
                    :title="
                      parameterGroupDisabledReason(selectedParameterGroup)
                    "
                    danger
                    size="small"
                    type="link"
                    @click="handleDeleteItem(record)"
                    >{{ pc('delete') }}</Button
                  >
                </Space>
              </template>
            </template>
          </Table>
        </template>
        <Empty
          v-else
          class="workspace-empty"
          :description="pc('emptyContext')"
        />
      </main>
    </div>

    <Modal
      v-model:open="entityModalOpen"
      :confirm-loading="entitySubmitting"
      :title="entityModalTitle"
      width="620px"
      @ok="handleEntitySubmit"
    >
      <Form ref="entityFormRef" :label-col="{ span: 5 }" :model="entityForm">
        <template v-if="entityKind === 'type'">
          <Form.Item
            :label="pc('typeName')"
            name="modelTypeName"
            :rules="[{ required: true, message: pc('requiredTypeName') }]"
          >
            <Input v-model:value="entityForm.modelTypeName" />
          </Form.Item>
          <Form.Item
            :label="pc('useCase')"
            name="useCase"
            :rules="[{ required: true }]"
          >
            <Select
              v-model:value="entityForm.useCase"
              :options="MODEL_TYPE_USE_CASE_OPTIONS"
            />
          </Form.Item>
          <Form.Item :label="pc('typeDescription')" name="typeDesc">
            <Input.TextArea v-model:value="entityForm.typeDesc" :rows="4" />
          </Form.Item>
        </template>
        <template v-else>
          <Form.Item
            :label="pc('modelName')"
            name="modelName"
            :rules="[{ required: true, message: pc('requiredModelName') }]"
          >
            <Input v-model:value="entityForm.modelName" />
          </Form.Item>
          <Form.Item
            :label="pc('deviceType')"
            name="modelTypeId"
            :rules="[{ required: true, message: pc('requiredDeviceType') }]"
          >
            <Select
              v-model:value="entityForm.modelTypeId"
              :options="
                modelTypeItems.map((item) => ({
                  label: item.modelTypeName,
                  value: item.id,
                }))
              "
            />
          </Form.Item>
          <Form.Item :label="pc('modelDescription')" name="description">
            <Input.TextArea v-model:value="entityForm.description" :rows="3" />
          </Form.Item>
          <Form.Item :label="pc('remark')" name="remark">
            <Input.TextArea v-model:value="entityForm.remark" :rows="3" />
          </Form.Item>
        </template>
      </Form>
    </Modal>

    <Drawer
      v-model:open="itemDrawerOpen"
      destroy-on-close
      :title="itemDrawerTitle"
      width="620"
    >
      <Alert
        class="mb-4"
        :message="`${selectedModel?.modelName ?? pc('deviceModel')} / ${selectedParameterGroup?.groupName ?? pc('groupName')}`"
        show-icon
        type="info"
      />
      <Alert
        v-if="itemMode === 'edit'"
        class="mb-4"
        :message="pc('keyChangeWarning')"
        show-icon
        type="warning"
      />
      <Form
        ref="itemFormRef"
        :disabled="itemReadonly"
        :label-col="{ span: 6 }"
        :model="itemForm"
      >
        <Form.Item
          :label="pc('targetGroup')"
          name="parameterGroupId"
          :rules="[{ required: true, message: pc('requiredGroup') }]"
        >
          <Select
            v-model:value="itemForm.parameterGroupId"
            :disabled="itemMode !== 'copy'"
            :options="itemTargetGroupOptions"
          />
        </Form.Item>
        <Form.Item
          :label="pc('parameterName')"
          name="parameterName"
          :rules="[{ required: true, message: pc('requiredParameterName') }]"
        >
          <Input v-model:value="itemForm.parameterName" />
        </Form.Item>
        <Form.Item
          :label="pc('parameterKey')"
          name="parameterKey"
          :rules="[
            { required: true, message: pc('requiredParameterKey') },
            {
              pattern: /^[A-Za-z][A-Za-z0-9_.-]*$/,
              message: pc('invalidParameterKey'),
            },
          ]"
        >
          <Input v-model:value="itemForm.parameterKey" />
        </Form.Item>
        <Form.Item
          :label="pc('valueType')"
          name="valueType"
          :rules="[{ required: true }]"
        >
          <Select
            v-model:value="itemForm.valueType"
            :options="VALUE_TYPE_OPTIONS"
            @change="handleValueTypeChange"
          />
        </Form.Item>
        <Form.Item :label="pc('defaultValue')" name="defaultValue">
          <Select
            v-if="itemForm.valueType === 'BOOL'"
            v-model:value="itemForm.defaultValue"
            allow-clear
            :options="[
              { label: 'true', value: 'true' },
              { label: 'false', value: 'false' },
            ]"
          />
          <Input.TextArea
            v-else-if="itemForm.valueType === 'JSON'"
            v-model:value="itemForm.defaultValue"
            :rows="7"
          />
          <Input
            v-else
            v-model:value="itemForm.defaultValue"
            :type="itemForm.valueType === 'NUMBER' ? 'number' : 'text'"
          />
        </Form.Item>
        <Form.Item
          :label="pc('constraintType')"
          name="constraintType"
          :rules="[{ required: true }]"
        >
          <Select
            v-model:value="itemForm.constraintType"
            :options="availableConstraintTypeOptions"
          />
        </Form.Item>
        <Form.Item
          v-if="itemForm.constraintType === 'RANGE'"
          :label="pc('range')"
          required
        >
          <Space.Compact block>
            <InputNumber
              v-model:value="constraintForm.min"
              class="w-full"
              :placeholder="pc('minimum')"
            />
            <InputNumber
              v-model:value="constraintForm.max"
              class="w-full"
              :placeholder="pc('maximum')"
            />
          </Space.Compact>
        </Form.Item>
        <Form.Item
          v-if="itemForm.constraintType === 'LENGTH'"
          :label="pc('maxLength')"
          required
        >
          <InputNumber
            v-model:value="constraintForm.maxLength"
            :min="1"
            :precision="0"
          />
        </Form.Item>
        <Form.Item :label="pc('unit')" name="unit"
          ><Input v-model:value="itemForm.unit"
        /></Form.Item>
        <Form.Item :label="pc('required')" name="required"
          ><Switch v-model:checked="itemForm.required"
        /></Form.Item>
        <Form.Item :label="pc('remark')" name="remark"
          ><Input.TextArea v-model:value="itemForm.remark" :rows="3"
        /></Form.Item>
      </Form>
      <template #extra>
        <Space>
          <Button @click="itemDrawerOpen = false">{{
            itemReadonly ? pc('close') : pc('cancel')
          }}</Button>
          <Button
            v-if="!itemReadonly"
            :loading="itemSubmitting"
            type="primary"
            @click="() => handleItemSubmit()"
          >
            {{ itemMode === 'copy' ? pc('copyParameter') : pc('save') }}
          </Button>
          <Button
            v-if="itemMode === 'create'"
            :loading="itemSubmitting"
            @click="handleItemSubmit(true)"
          >
            {{ pc('saveAndContinue') }}
          </Button>
        </Space>
      </template>
    </Drawer>

    <Modal
      v-model:open="bindOpen"
      :confirm-loading="bindSubmitting"
      :ok-button-props="{ disabled: bindSelectedIds.length === 0 }"
      :title="pc('bindExistingGroup')"
      width="920px"
      @ok="handleBindGroups"
    >
      <Input
        v-model:value="bindSearch"
        allow-clear
        class="mb-3"
        :placeholder="pc('searchGroup')"
      >
        <template #prefix><IconifyIcon icon="lucide:search" /></template>
      </Input>
      <Table
        :columns="bindColumns"
        :data-source="filteredBindGroups"
        :loading="bindLoading"
        :pagination="{ pageSize: 10 }"
        :row-selection="{
          selectedRowKeys: bindSelectedIds,
          getCheckboxProps: (record: ParameterGroupRow) => ({
            disabled: !!bindDisabledReason(record),
          }),
          onChange: handleBindSelectionChange,
        }"
        row-key="id"
        size="small"
      >
        <template #bodyCell="{ column, record, text }">
          <template v-if="column.key === 'groupName'">
            {{ text }}
            <Tag v-if="bindDisabledReason(record)">{{
              bindDisabledReason(record)
            }}</Tag>
          </template>
          <template v-else-if="column.key === 'shareScope'">{{
            resolveShareScope(text)
          }}</template>
        </template>
      </Table>
    </Modal>

    <Drawer
      v-model:open="groupDrawerOpen"
      destroy-on-close
      :title="groupDrawerTitle"
      width="580"
    >
      <Alert
        class="mb-4"
        :message="
          groupEditingId
            ? pc('editGroupImpact', { count: groupEditingModelCount })
            : pc('createBindHint')
        "
        show-icon
        :type="
          groupEditingId && groupEditingModelCount > 1 ? 'warning' : 'info'
        "
      />
      <Form ref="groupFormRef" :label-col="{ span: 6 }" :model="groupForm">
        <Form.Item
          :label="pc('groupCode')"
          name="groupCode"
          :rules="[{ required: true, message: pc('requiredGroupCode') }]"
        >
          <Input v-model:value="groupForm.groupCode" />
        </Form.Item>
        <Form.Item
          :label="pc('groupName')"
          name="groupName"
          :rules="[{ required: true, message: pc('requiredGroupName') }]"
        >
          <Input v-model:value="groupForm.groupName" />
        </Form.Item>
        <Form.Item
          :label="pc('groupType')"
          name="groupType"
          :rules="[{ required: true }]"
        >
          <Select
            v-model:value="groupForm.groupType"
            :options="GROUP_TYPE_OPTIONS"
          />
        </Form.Item>
        <Form.Item
          :label="pc('shareScope')"
          name="shareScope"
          :rules="[{ required: true }]"
        >
          <Select
            v-model:value="groupForm.shareScope"
            :options="SHARE_SCOPE_OPTIONS"
          />
        </Form.Item>
        <Form.Item :label="pc('editable')" name="editable"
          ><Switch v-model:checked="groupForm.editable"
        /></Form.Item>
        <Form.Item :label="pc('version')" name="version"
          ><InputNumber v-model:value="groupForm.version" :min="1"
        /></Form.Item>
        <Form.Item :label="pc('description')" name="description"
          ><Input.TextArea v-model:value="groupForm.description" :rows="4"
        /></Form.Item>
      </Form>
      <template #extra>
        <Space>
          <Button @click="groupDrawerOpen = false">{{ pc('cancel') }}</Button>
          <Button
            :loading="groupSubmitting"
            type="primary"
            @click="handleCreateAndBind"
            >{{
              groupEditingId ? pc('saveChanges') : pc('createAndBindGroup')
            }}</Button
          >
        </Space>
      </template>
    </Drawer>

    <Drawer
      v-model:open="copyDrawerOpen"
      destroy-on-close
      :title="pc('copyGroup')"
      width="540"
    >
      <Alert
        class="mb-4"
        :message="pc('copyTransactionWarning')"
        show-icon
        type="warning"
      />
      <Form ref="copyFormRef" :label-col="{ span: 6 }" :model="copyForm">
        <Form.Item
          :label="pc('newCode')"
          name="groupCode"
          :rules="[{ required: true, message: pc('requiredNewCode') }]"
        >
          <Input v-model:value="copyForm.groupCode" />
        </Form.Item>
        <Form.Item
          :label="pc('newName')"
          name="groupName"
          :rules="[{ required: true, message: pc('requiredNewName') }]"
        >
          <Input v-model:value="copyForm.groupName" />
        </Form.Item>
        <Form.Item :label="pc('version')" name="version"
          ><InputNumber v-model:value="copyForm.version" :min="1"
        /></Form.Item>
        <Form.Item :label="pc('copyParameter')" name="copyItems"
          ><Checkbox v-model:checked="copyForm.copyItems">{{
            pc('copyAllItems')
          }}</Checkbox></Form.Item
        >
      </Form>
      <template #extra>
        <Space>
          <Button @click="copyDrawerOpen = false">{{ pc('cancel') }}</Button>
          <Button
            :loading="copySubmitting"
            type="primary"
            @click="handleCopyGroup"
            >{{ pc('copyAndBind') }}</Button
          >
        </Space>
      </template>
    </Drawer>

    <Modal
      v-model:open="associationOpen"
      :footer="null"
      :title="pc('association')"
      width="min(1200px, 84vw)"
    >
      <div v-if="selectedParameterGroup" class="association-summary">
        <div>
          <strong>{{ selectedParameterGroup.groupName }}</strong
          ><span>{{ selectedParameterGroup.groupCode }}</span>
        </div>
        <Space>
          <Tag color="blue">{{
            pc('modelsSummary', { count: associationModels.length })
          }}</Tag>
          <Tag color="green">{{
            pc('devicesSummary', { count: associationDevices.length })
          }}</Tag>
          <Tag color="orange">{{
            pc('groupsSummary', { count: associationGroups.length })
          }}</Tag>
        </Space>
      </div>
      <Tabs v-model:active-key="associationTab">
        <Tabs.TabPane key="models" :tab="pc('associatedModels')">
          <Table
            :columns="associationModelColumns"
            :data-source="associationModels"
            :loading="associationLoading"
            row-key="id"
            size="small"
          />
        </Tabs.TabPane>
        <Tabs.TabPane key="devices" :tab="pc('deviceInstances')">
          <Table
            :columns="associationDeviceColumns"
            :data-source="associationDevices"
            :loading="associationLoading"
            row-key="id"
            size="small"
          />
        </Tabs.TabPane>
        <Tabs.TabPane key="groups" :tab="pc('groupDistribution')">
          <Table
            :columns="associationGroupColumns"
            :data-source="associationGroups"
            :loading="associationLoading"
            row-key="id"
            size="small"
          />
        </Tabs.TabPane>
      </Tabs>
    </Modal>
  </Page>
</template>

<style scoped>
.parameter-management {
  display: grid;
  grid-template-columns: 340px minmax(0, 1fr);
  gap: 12px;
  height: 100%;
  min-height: 620px;
}

.parameter-management--collapsed {
  grid-template-columns: 48px minmax(0, 1fr);
}

.context-workbench,
.parameter-workspace {
  min-width: 0;
  height: 100%;
  background: hsl(var(--background));
  border: 1px solid hsl(var(--border));
  border-radius: 8px;
}

.context-workbench {
  position: relative;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.collapse-button {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 2;
}

.context-section {
  display: flex;
  flex: none;
  flex-direction: column;
  min-height: 0;
  padding: 14px;
  overflow: hidden;
}

.group-section {
  flex: 1;
}

.section-header,
.workspace-header {
  display: flex;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
}

.section-header {
  min-height: 34px;
  padding-right: 28px;
  margin-bottom: 10px;
}

.section-title,
.workspace-title {
  font-size: 15px;
  font-weight: 600;
  color: hsl(var(--foreground));
}

.section-subtitle,
.workspace-eyebrow,
.workspace-meta {
  margin-top: 2px;
  font-size: 12px;
  color: hsl(var(--muted-foreground));
}

.context-search {
  margin-bottom: 10px;
}

.tree-scroll,
.group-list {
  min-height: 0;
  overflow: auto;
}

.model-tree-node {
  display: flex;
  gap: 4px;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.model-tree-node__actions {
  display: flex;
  flex: none;
  opacity: 0;
}

.model-tree-node:hover .model-tree-node__actions,
.model-tree-node:focus-within .model-tree-node__actions {
  opacity: 1;
}

.vertical-resizer {
  display: flex;
  flex: none;
  align-items: center;
  justify-content: center;
  height: 9px;
  cursor: row-resize;
  border-block: 1px solid hsl(var(--border));
}

.vertical-resizer span {
  width: 34px;
  height: 3px;
  background: hsl(var(--muted-foreground) / 35%);
  border-radius: 2px;
}

.group-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.group-item {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 4px;
  align-items: center;
  width: 100%;
  padding: 9px 10px;
  color: hsl(var(--foreground));
  text-align: left;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 6px;
}

.group-item__content {
  min-width: 0;
  padding: 0;
  color: inherit;
  text-align: left;
  background: transparent;
  border: 0;
}

.group-item__actions {
  display: flex;
  align-items: center;
  opacity: 0;
  transition: opacity 120ms ease;
}

.group-item:hover .group-item__actions,
.group-item:focus-within .group-item__actions,
.group-item--selected .group-item__actions {
  opacity: 1;
}

.group-item:hover {
  background: hsl(var(--accent) / 45%);
}

.group-item--selected {
  background: hsl(var(--primary) / 10%);
  border-color: hsl(var(--primary) / 35%);
}

.group-item__main,
.group-item__meta {
  display: flex;
  gap: 8px;
  align-items: center;
  justify-content: space-between;
}

.group-item__name {
  overflow: hidden;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.group-item__count,
.group-item__meta {
  font-size: 12px;
  color: hsl(var(--muted-foreground));
}

.group-item__meta {
  justify-content: flex-start;
  margin-top: 4px;
}

.section-placeholder,
.workspace-empty {
  display: grid;
  flex: 1;
  place-items: center;
  color: hsl(var(--muted-foreground));
}

.collapsed-context {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: center;
  padding-top: 52px;
  font-size: 12px;
  color: hsl(var(--muted-foreground));
  writing-mode: vertical-rl;
}

.parameter-workspace {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 16px;
  overflow: hidden;
}

.workspace-header {
  flex: none;
  min-height: 58px;
  padding-bottom: 12px;
  border-bottom: 1px solid hsl(var(--border));
}

.workspace-title {
  overflow: hidden;
  font-size: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.parameter-toolbar {
  display: grid;
  grid-template-columns: minmax(200px, 1fr) 140px 140px auto;
  gap: 8px;
  align-self: start;
}

.parameter-table {
  min-height: 0;
  overflow: hidden;
}

.parameter-table :deep(.parameter-remark-cell) {
  min-width: 260px;
  white-space: nowrap;
}

.min-w-0 {
  min-width: 0;
}

.text-ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.association-summary {
  display: flex;
  gap: 16px;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  margin-bottom: 8px;
  background: hsl(var(--muted) / 45%);
  border: 1px solid hsl(var(--border));
  border-radius: 6px;
}

.association-summary > div {
  display: flex;
  gap: 10px;
  align-items: baseline;
}

.association-summary span {
  font-size: 12px;
  color: hsl(var(--muted-foreground));
}

@media (max-width: 900px) {
  .parameter-management,
  .parameter-management--collapsed {
    grid-template-columns: 1fr;
    height: auto;
  }

  .context-workbench {
    height: 520px;
  }

  .parameter-workspace {
    min-height: 620px;
  }
}
</style>
