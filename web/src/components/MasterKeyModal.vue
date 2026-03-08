<template>
  <n-modal
    :show="show"
    preset="card"
    :title="t('auth.title')"
    :mask-closable="false"
    :closable="false"
    style="max-width: 480px"
    class="auth-modal"
  >
    <n-space vertical size="large">
      <div class="lang-switch">
        <n-dropdown :options="languageOptions" @select="onSelectLanguage">
          <n-button quaternary size="small">
            <template #icon>
              <n-icon :component="LanguageOutline" />
            </template>
            {{ locale === "zh-CN" ? "CN" : "EN" }}
          </n-button>
        </n-dropdown>
      </div>
      <div class="auth-header">
        <n-icon size="48" :component="LockClosedOutline" class="auth-icon" />
        <div class="auth-title">{{ t("auth.welcome") }}</div>
        <div class="auth-subtitle">
          {{ authSubtitle }}
        </div>
      </div>

      <n-alert v-if="error" type="error" closable class="error-alert">
        {{ error }}
      </n-alert>

      <n-form-item :label="usernameLabel" label-placement="top">
        <n-input
          v-model:value="username"
          :placeholder="usernamePlaceholder"
          size="large"
          @keyup.enter="onSubmit"
          autofocus
        >
          <template #prefix>
            <n-icon :component="PersonOutline" />
          </template>
        </n-input>
      </n-form-item>

      <n-form-item :label="passwordLabel" label-placement="top">
        <n-input
          v-model:value="password"
          type="password"
          :placeholder="passwordPlaceholder"
          show-password-on="mousedown"
          size="large"
          @keyup.enter="onSubmit"
        >
          <template #prefix>
            <n-icon :component="KeyOutline" />
          </template>
        </n-input>
      </n-form-item>

      <n-button
        type="primary"
        size="large"
        block
        :loading="loading"
        :disabled="!username.trim() || !password.trim()"
        @click="onSubmit"
      >
        {{ t("auth.accessDashboard") }}
      </n-button>

      <div class="auth-footer">
        {{ authFooter }}
      </div>
    </n-space>
  </n-modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import {
  NAlert,
  NButton,
  NDropdown,
  NFormItem,
  NIcon,
  NInput,
  NModal,
  NSpace,
} from "naive-ui";
import {
  KeyOutline,
  LanguageOutline,
  LockClosedOutline,
  PersonOutline,
} from "@vicons/ionicons5";
import { locale, setLocale, t } from "../i18n";

const props = defineProps<{
  show: boolean;
  initialUsername?: string;
  error?: string;
  loading?: boolean;
}>();

const emit = defineEmits<{
  (e: "submit", payload: { username: string; password: string }): void;
}>();

const username = ref(props.initialUsername ?? "admin");
const password = ref("");

const usernameLabel = computed(() => t("auth.usernameLabel"));
const usernamePlaceholder = computed(() => t("auth.usernamePlaceholder"));
const passwordLabel = computed(() => t("auth.passwordLabel"));
const passwordPlaceholder = computed(() => t("auth.passwordPlaceholder"));
const authSubtitle = computed(() => t("auth.subtitle"));
const authFooter = computed(() => t("auth.footer"));

const languageOptions = [
  { label: "English", key: "en" },
  { label: "Chinese", key: "zh-CN" },
];

function onSelectLanguage(key: string | number) {
  if (key === "en" || key === "zh-CN") {
    setLocale(key);
  }
}

watch(
  () => props.initialUsername,
  (v) => {
    if (typeof v === "string") username.value = v;
  }
);

watch(
  () => props.show,
  (show) => {
    if (show) {
      password.value = "";
    }
  }
);

function onSubmit() {
  if (!username.value.trim() || !password.value.trim()) return;
  emit("submit", {
    username: username.value.trim(),
    password: password.value,
  });
}
</script>

<style scoped>
.auth-modal {
  border-radius: 20px;
}

.lang-switch {
  display: flex;
  justify-content: flex-end;
}

.auth-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.auth-icon {
  color: #18a058;
  margin-bottom: 8px;
}

.auth-title {
  font-size: 22px;
  font-weight: 700;
}

.auth-subtitle {
  color: #888;
  text-align: center;
  font-size: 14px;
}

.error-alert {
  border-radius: 8px;
}

.auth-footer {
  text-align: center;
  color: #bbb;
  font-size: 12px;
  margin-top: 8px;
}
</style>
