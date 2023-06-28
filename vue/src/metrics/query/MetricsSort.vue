<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" :disabled="disabled" v-bind="attrs" v-on="on">
        Sort
      </v-btn>
    </template>

    <SearchableList :items="orders" @input="sortBy($event)"></SearchableList>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, PropType, shallowRef } from "vue";

// Composables
import { UseUql } from '@/use/uql'

// Components
import SearchableList from '@/components/SearchableList.vue'

export default defineComponent({
  name: 'MetricsSort',
  components: { SearchableList },

  props: {
    uql: {
      type: Object as PropType<UseUql>,
      required: true,
    },
    attrKeys: {
      type: Array as PropType<string[]>,
      required: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    sort: {
      type: String,
      default: 'desc',
    },
  },

  setup(props) {
    const menu = shallowRef(false)
    const orders = ['asc', 'desc']

    function sortBy(order: string) {
      props.uql.order = order
      menu.value = false
    }

    return {
      menu,
      orders,
      sortBy,
    }
  },
})
</script>

<style lang="scss" scoped></style>
