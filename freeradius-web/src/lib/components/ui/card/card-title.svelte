<script lang="ts">
	import { cn } from '$lib/utils.js';
	let klass: string | undefined = undefined;
	export { klass as class };
</script>

<h3 class={cn('font-semibold leading-none tracking-tight', klass)} {...$$restProps}>
	<slot />
</h3>
