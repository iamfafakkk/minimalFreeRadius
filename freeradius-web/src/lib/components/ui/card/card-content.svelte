<script lang="ts">
	import { cn } from '$lib/utils.js';
	let klass: string | undefined = undefined;
	export { klass as class };
</script>

<div class={cn('p-6 pt-0', klass)} {...$$restProps}>
	<slot />
</div>
