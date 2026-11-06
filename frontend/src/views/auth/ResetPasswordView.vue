<template>
  <AuthLayout>
    <div class="space-y-6">
      <!-- Title -->
      <div class="text-center">
        <h2 class="text-2xl font-light tracking-wide text-[#3d3d3d] dark:text-[#ede7e0]">
          {{ t('auth.resetPasswordTitle') }}
        </h2>
        <p class="mt-2 text-sm font-light text-[#756554] dark:text-[#c9a88c]">
          {{ t('auth.resetPasswordHint') }}
        </p>
      </div>

      <!-- Invalid Link State -->
      <div v-if="isInvalidLink" class="space-y-6">
        <div class="rounded-lg border border-[#a89279]/30 bg-[#a89279]/10 p-6 dark:border-[#c9a88c]/25 dark:bg-[#c9a88c]/10">
          <div class="flex flex-col items-center gap-4 text-center">
            <div class="flex h-12 w-12 items-center justify-center rounded-sm bg-[#a89279]/15 dark:bg-[#c9a88c]/15">
              <Icon name="exclamationCircle" size="lg" class="text-[#756554] dark:text-[#c9a88c]" />
            </div>
            <div>
              <h3 class="text-lg font-light tracking-wide text-[#3d3d3d] dark:text-[#ede7e0]">
                {{ t('auth.invalidResetLink') }}
              </h3>
              <p class="mt-2 text-sm font-light text-[#756554] dark:text-[#c9a88c]">
                {{ t('auth.invalidResetLinkHint') }}
              </p>
            </div>
          </div>
        </div>

        <div class="text-center">
          <router-link
            to="/forgot-password"
            class="inline-flex items-center gap-2 font-light lowercase tracking-wide text-[#756554] transition-colors duration-500 hover:text-[#3d3d3d] dark:text-[#c9a88c] dark:hover:text-[#ede7e0]"
          >
            {{ t('auth.requestNewResetLink') }}
          </router-link>
        </div>
      </div>

      <!-- Success State -->
      <div v-else-if="isSuccess" class="space-y-6">
        <div class="rounded-lg border border-[#5a7a6b]/25 bg-[#5a7a6b]/10 p-6 dark:border-[#9ab5a6]/20 dark:bg-[#9ab5a6]/10">
          <div class="flex flex-col items-center gap-4 text-center">
            <div class="flex h-12 w-12 items-center justify-center rounded-sm bg-[#5a7a6b]/15 dark:bg-[#9ab5a6]/15">
              <Icon name="checkCircle" size="lg" class="text-[#4f6b5e] dark:text-[#9ab5a6]" />
            </div>
            <div>
              <h3 class="text-lg font-light tracking-wide text-[#3d3d3d] dark:text-[#ede7e0]">
                {{ t('auth.passwordResetSuccess') }}
              </h3>
              <p class="mt-2 text-sm font-light text-[#756554] dark:text-[#c9a88c]">
                {{ t('auth.passwordResetSuccessHint') }}
              </p>
            </div>
          </div>
        </div>

        <div class="text-center">
          <router-link
            to="/login"
            class="btn btn-primary inline-flex items-center gap-2"
          >
            <Icon name="login" size="md" />
            {{ t('auth.signIn') }}
          </router-link>
        </div>
      </div>

      <!-- Form State -->
      <form v-else @submit.prevent="handleSubmit" class="space-y-5">
        <!-- Email (readonly) -->
        <div>
          <label for="email" class="input-label">
            {{ t('auth.emailLabel') }}
          </label>
          <div class="relative">
            <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
              <Icon name="mail" size="md" class="text-[#756554] dark:text-[#c9a88c]" />
            </div>
            <input
              id="email"
              :value="email"
              type="email"
              readonly
              disabled
              class="input pl-11 bg-[#ede7e0] dark:bg-[#242321]"
            />
          </div>
        </div>

        <!-- New Password Input -->
        <div>
          <label for="password" class="input-label">
            {{ t('auth.newPassword') }}
          </label>
          <div class="relative">
            <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
              <Icon name="lock" size="md" class="text-[#756554] dark:text-[#c9a88c]" />
            </div>
            <input
              id="password"
              v-model="formData.password"
              :type="showPassword ? 'text' : 'password'"
              required
              autocomplete="new-password"
              :disabled="isLoading"
              class="input pl-11 pr-11"
              :class="{ 'input-error': errors.password }"
              :placeholder="t('auth.newPasswordPlaceholder')"
            />
            <button
              type="button"
              @click="showPassword = !showPassword"
              class="absolute inset-y-0 right-0 flex items-center pr-3.5 text-[#756554] transition-colors duration-500 hover:text-[#3d3d3d] dark:text-[#c9a88c] dark:hover:text-[#ede7e0]"
            >
              <Icon v-if="showPassword" name="eyeOff" size="md" />
              <Icon v-else name="eye" size="md" />
            </button>
          </div>
        </div>

        <!-- Confirm Password Input -->
        <div>
          <label for="confirmPassword" class="input-label">
            {{ t('auth.confirmPassword') }}
          </label>
          <div class="relative">
            <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
              <Icon name="lock" size="md" class="text-[#756554] dark:text-[#c9a88c]" />
            </div>
            <input
              id="confirmPassword"
              v-model="formData.confirmPassword"
              :type="showConfirmPassword ? 'text' : 'password'"
              required
              autocomplete="new-password"
              :disabled="isLoading"
              class="input pl-11 pr-11"
              :class="{ 'input-error': errors.confirmPassword }"
              :placeholder="t('auth.confirmPasswordPlaceholder')"
            />
            <button
              type="button"
              @click="showConfirmPassword = !showConfirmPassword"
              class="absolute inset-y-0 right-0 flex items-center pr-3.5 text-[#756554] transition-colors duration-500 hover:text-[#3d3d3d] dark:text-[#c9a88c] dark:hover:text-[#ede7e0]"
            >
              <Icon v-if="showConfirmPassword" name="eyeOff" size="md" />
              <Icon v-else name="eye" size="md" />
            </button>
          </div>
        </div>

        <!-- Submit Button -->
        <button
          type="submit"
          :disabled="isLoading"
          class="btn btn-primary w-full"
        >
          <svg
            v-if="isLoading"
            class="-ml-1 mr-2 h-4 w-4 animate-spin text-[#f5f0eb] dark:text-[#2f2d2a]"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          <Icon v-else name="checkCircle" size="md" class="mr-2" />
          {{ isLoading ? t('auth.resettingPassword') : t('auth.resetPassword') }}
        </button>
      </form>
    </div>

    <!-- Footer -->
    <template #footer>
      <p class="font-light text-[#756554] dark:text-[#c9a88c]">
        {{ t('auth.rememberedPassword') }}
        <router-link
          to="/login"
          class="font-light lowercase tracking-wide text-[#756554] transition-colors duration-500 hover:text-[#3d3d3d] dark:text-[#c9a88c] dark:hover:text-[#ede7e0]"
        >
          {{ t('auth.signIn') }}
        </router-link>
      </p>
    </template>
  </AuthLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { resetPassword } from '@/api/auth'

const { t } = useI18n()

// ==================== Router & Stores ====================

const route = useRoute()
const appStore = useAppStore()

// ==================== State ====================

const isLoading = ref<boolean>(false)
const isSuccess = ref<boolean>(false)
const errorMessage = ref<string>('')
const showPassword = ref<boolean>(false)
const showConfirmPassword = ref<boolean>(false)

// URL parameters
const email = ref<string>('')
const token = ref<string>('')

const formData = reactive({
  password: '',
  confirmPassword: ''
})

const errors = reactive({
  password: '',
  confirmPassword: ''
})

const validationToastMessage = computed(
  () => errors.password || errors.confirmPassword || ''
)

watch(validationToastMessage, (value, previousValue) => {
  if (value && value !== previousValue) {
    appStore.showError(value)
  }
})

// Check if the reset link is valid (has email and token)
const isInvalidLink = computed(() => !email.value || !token.value)

// ==================== Lifecycle ====================

onMounted(() => {
  // Get email and token from URL query parameters
  email.value = (route.query.email as string) || ''
  token.value = (route.query.token as string) || ''

  if (!email.value || !token.value) {
    appStore.showError(t('auth.invalidResetLink'))
  }
})

// ==================== Validation ====================

function validateForm(): boolean {
  errors.password = ''
  errors.confirmPassword = ''

  let isValid = true

  // Password validation
  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
    isValid = false
  }

  // Confirm password validation
  if (!formData.confirmPassword) {
    errors.confirmPassword = t('auth.confirmPasswordRequired')
    isValid = false
  } else if (formData.password !== formData.confirmPassword) {
    errors.confirmPassword = t('auth.passwordsDoNotMatch')
    isValid = false
  }

  return isValid
}

// ==================== Form Handlers ====================

async function handleSubmit(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    await resetPassword({
      email: email.value,
      token: token.value,
      new_password: formData.password
    })

    isSuccess.value = true
    appStore.showSuccess(t('auth.passwordResetSuccess'))
  } catch (error: unknown) {
    const err = error as { message?: string; response?: { data?: { detail?: string; code?: string } } }

    // Check for invalid/expired token error
    if (err.response?.data?.code === 'INVALID_RESET_TOKEN') {
      errorMessage.value = t('auth.invalidOrExpiredToken')
    } else if (err.response?.data?.detail) {
      errorMessage.value = err.response.data.detail
    } else if (err.message) {
      errorMessage.value = err.message
    } else {
      errorMessage.value = t('auth.resetPasswordFailed')
    }

    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.5s ease-in-out;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
