<template>
  <v-menu v-model="menu" offset-y :close-on-content-click="false">
    <template #activator="{ on, attrs }">
      <v-btn text class="v-btn--filter" :disabled="disabled" v-bind="attrs" v-on="on">
        topN
      </v-btn>
    </template>

    <SearchableList :items="topNList" @input="setTopN($event)"></SearchableList>
  </v-menu>
</template>

<script lang="ts">
import { defineComponent, PropType, shallowRef } from "vue";

// Composables
import { UseUql } from '@/use/uql'

// Components
import SearchableList from '@/components/SearchableList.vue'

export default defineComponent({
  name: 'MetricsTopN',
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
    const topNList = ['1', '2', '3', '4', '5', '6', '7', '8', '9', '10', '15', '20', '30', '50', '75', '100']

    function setTopN(topN: string) {
      props.uql.topN = parseInt(topN)
      menu.value = false
    }

    return {
      menu,
      topNList,
      setTopN,
    }
  },
})
</script>

<style lang="scss" scoped></style>
