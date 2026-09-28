<script lang="ts">
	import { cn } from '$lib/utils.js';
	let klass: string | undefined = undefined;
	export { klass as class };
</script>

<thead class={cn('[&_tr]:border-b', klass)} {...$$restProps}>
	<slot />
</thead>
