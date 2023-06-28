<template>
  <div class="d-flex flex-wrap justify-center text-caption" :class="`flex-${direction}`">
    <div
      v-for="item in sortedTimeseries"
      :key="item.name"
      class="mx-2 d-flex align-center cursor-pointer"
      :class="{ 'text--secondary': !isSelected(item) }"
      @mouseenter="$emit('hover:item', { item: item, hover: true })"
      @mouseleave="$emit('hover:item', { item: item, hover: false })"
      @click.exact.stop="toggle(item)"
      @click.meta.stop="select(item)"
    >
      <v-icon size="16" :color="isSelected(item) ? item.color : 'grey'" class="mr-1"
        >mdi-circle</v-icon
      >
      <div>{{ truncateMiddle(item.name, 40) }}</div>
      <div v-for="metric in values" :key="metric" class="ml-1 font-weight-medium">
        <XNum :value="item[metric]" :unit="item.unit" short :title="`${metric}: {0}`" />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, shallowRef, watch, PropType, computed } from "vue";

// Utilities
import { StyledTimeseries, LegendValue } from '@/metrics/types'
import { truncateMiddle } from '@/util/string'

export default defineComponent({
  name: 'ChartLegendList',

  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    timeseries: {
      type: Array as PropType<StyledTimeseries[]>,
      required: true,
    },
    values: {
      type: Array as PropType<LegendValue[]>,
      default: () => [LegendValue.Avg, LegendValue.Last, LegendValue.Min, LegendValue.Max],
    },
    direction: {
      type: String,
      default: 'row',
    },
    order: {
      type: String,
      default: 'desc',
    },
    topN: {
      type: Number,
      default: 5,
    },
  },

  setup(props, ctx) {
    const timeseries = shallowRef<StyledTimeseries[]>(props.timeseries)
    const sortedTimeseries = computed(() => {
      return sortTimeseries(timeseries.value, props.order, props.topN)
    })
    watch(
      () => props.timeseries,
      (tss) => (timeseries.value = tss),
      { immediate: true },
    )
    watch(
      () => props.order,
      (order) => {
        timeseries.value = sortTimeseries(props.timeseries, order, props.topN)
      },
      {immediate: true},
    )
    watch(
      () => props.topN,
      (topN) => {
        timeseries.value = sortTimeseries(props.timeseries, props.order, topN)
      },
      {immediate: true},
    )
    watch(timeseries, (selectedTimeseries) => ctx.emit('current-items', selectedTimeseries))

    function toggle(ts: StyledTimeseries) {
      const items = timeseries.value.slice()
      const index = items.findIndex((item) => item.id === ts.id)
      if (index >= 0) {
        items.splice(index, 1)
      } else {
        items.push(ts)
      }
      timeseries.value = items
    }

    function select(ts: StyledTimeseries) {
      const items = timeseries.value.slice()
      if (items.length == props.timeseries.length) {
        timeseries.value = [ts]
        return
      }
      const index = items.findIndex((item) => item.id === ts.id)
      if (index < 0) {
        items.push(ts)
      } else {
        items.splice(index, 1)
      }
      if (items.length == 0) {
        timeseries.value = props.timeseries
      } else {
        timeseries.value = items
      }
    }

    function isSelected(ts: StyledTimeseries): boolean {
      const index = timeseries.value.findIndex((item) => item.id === ts.id)
      return index >= 0
    }

    function sortTimeseries(timeseries: StyledTimeseries[], order: string, topN: number): StyledTimeseries[] {
      const key = props.values[0]
      if (order == 'desc') {
        return props.timeseries
          .sort(function (a, b) {
            return b[key] - a[key]
          })
          .slice(0, topN)
      } else {
        return props.timeseries
          .slice(0, topN)
          .sort(function (a, b) {
            return a[key] - b[key]
          })
          .slice(0, topN)
      }
    }

    return {
      selectedTimeseries: timeseries,
      toggle,
      select,
      isSelected,
      sortedTimeseries,

      truncateMiddle,
    }
  },
})
</script>

<style lang="scss" scoped></style>
